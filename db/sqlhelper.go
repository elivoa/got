/*
  Latest one.

 SQL Helper is a helper method in total filtering.

 Usage Examples:
  em.Select().Where("id", 5).QueryOne()
  em.Select("id","name").Where("type", "person").Query()
  ...
  em.Update().Where("id", 5).Exec(name, class, ...)
  em.Update().Exec(name, class, ..., id)
  em.Update("time").Where("id", 5, "person", 6).Exec(time)

  TODO Count();

>>DarkType
*/
package db

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/elivoa/got/logs"
	"github.com/elivoa/gxl"
	"strings"
)

// ________________________________________

var sqllogger = logs.Get("SQL:Print")

// Entities cache.
// TODO: thread safe? need lock?
var entities map[string]*Entity
var queryparserCache map[string]*QueryParser

func init() {
	entities = make(map[string]*Entity, 10)
	queryparserCache = make(map[string]*QueryParser)
}

// constants
var (
	ASC  = "asc"
	DESC = "desc"
)

func RegisterEntity(name string, entity *Entity) {
	if _, ok := entities[name]; ok {
		panic(fmt.Sprintf("DB: Register duplicated entities for %s", name))
	}
	entities[name] = entity
}

// --------------------------------------------------------------------------------
func NewQueryParser() *QueryParser {
	return &QueryParser{
		conditions: make([]*condition, 0),
	}
}

// ________________________________________________________________________________
// DAO Helper
type Entity struct {
	Table        string   // table name
	Alias        string   // from 'table' as Alias
	PK           string   // primary key field name
	Fields       []string // field names
	CreateFields []string // fields used in create things.
	UpdateFields []string // fields used in create things.
}

// TODO Cache queryParser here.
func (e *Entity) Create(queryName string) *QueryParser {
	parser := &QueryParser{
		e: e,
	}
	return parser
}

// 1st step: choose query type.
// 两种用法之一：先从select开始。然后添加where等条件。这种方法不再开发了，为了兼容考虑留下这些方法。
// 只有Select, Count, 参与这种。
func (e *Entity) Select(fields ...string) *QueryParser {
	return e.createQueryParser("select", fields...)
}

func (e *Entity) Insert(fields ...string) *QueryParser {
	return e.createQueryParser("insert", fields...)
}

func (e *Entity) Update(fields ...string) *QueryParser {
	return e.createQueryParser("update", fields...)
}

// alias is sql
func (e *Entity) RawQuery(sql string) *QueryParser {
	return e.createQueryParser("sql", sql)
}

func (e *Entity) Delete() *QueryParser {
	return e.createQueryParser("delete")
}

func (e *Entity) Count() *QueryParser {
	return e.createQueryParser("count")
}

// func (e *Entity) Close() *QueryParser {
// 	return e.createQueryParser("delete")
// }

// Create一个QueryParser,可以从Where条件开始，添加各种其他条件。最后调用select或者是count的一种builder方法。
func (e *Entity) NewQueryParser() *QueryParser {
	return &QueryParser{
		e:          e,
		conditions: make([]*condition, 0),
	}
}

// 兼容的方法
func (e *Entity) createQueryParser(operation string, fields ...string) *QueryParser {
	parser := &QueryParser{
		e:          e,
		operation:  operation,
		fields:     fields,
		conditions: make([]*condition, 0),
	}
	if nil != fields && len(fields) > 0 {
		parser.useCustomerFields = true
	}
	return parser
}

// TODO not used
func (e *Entity) NamedQuery(name string, createfunc func() *QueryParser) *QueryParser {
	cached, ok := queryparserCache[name]
	if !ok {
		cached = createfunc()
		queryparserCache[name] = cached
	}
	return cached

}

// ________________________________________________________________________________
// Query parser
//
type QueryParser struct {
	e          *Entity
	operation  string       // select, insert, update, insertorupdate, delete
	fields     []string     // selected fields, only select clause use this
	conditions []*condition // where 'id' = 1
	limit      *gxl.Int     // limit 4
	n          *gxl.Int     // limit 'limit','n'
	orderby    string       // TODO change to field [asc|desc]
	order      string       // asc | desc

	prepared          bool // status.
	useCustomerFields bool // only select clause use this

	sql    string        // generated sql
	values []interface{} // values in sequence used to inject into sql.
}

type condition struct {
	field  string
	values []interface{} // values, if only 1 value, use values[0]
	op     string        // and, or, andx, orx,
}

func (p *QueryParser) SetEntity(entity *Entity) *QueryParser {
	p.e = entity
	return p
}

func (p *QueryParser) Reset() *QueryParser {
	p.sql = ""
	p.values = []interface{}{}
	p.prepared = false
	return p
}

func (p *QueryParser) Fields(fields ...string) *QueryParser {
	p.useCustomerFields = true
	p.fields = fields
	return p
}

func (p *QueryParser) Where(conditions ...interface{}) *QueryParser {
	p.conditions = []*condition{}
	if len(conditions) == 0 {
	} else if len(conditions) == 2 {
		p.conditions = append(p.conditions, &condition{
			field:  conditions[0].(string),
			values: []interface{}{conditions[1]},
			op:     "and",
		})
	} else {
		panic("Where clouse only accept 0 or 2 parameters.")
	}
	// TODO
	return p
}

func (p *QueryParser) And(field string, values ...interface{}) *QueryParser {
	p.conditions = append(p.conditions, &condition{field: field, values: values, op: "and"})
	return p
}

func (p *QueryParser) AndRaw(sqlFragemnt string, values ...interface{}) *QueryParser {
	p.conditions = append(p.conditions, &condition{field: sqlFragemnt, values: values, op: "sql"})
	return p
}

// and (x1 or x2 or ...) name should be Orxp
func (p *QueryParser) Or(field string, values ...interface{}) *QueryParser {
	p.conditions = append(p.conditions, &condition{field: field, values: values, op: "or"})
	return p
}

// Now only support "and like".
func (p *QueryParser) Like(field string, values ...interface{}) *QueryParser {
	p.conditions = append(p.conditions, &condition{field: field, values: values, op: "like"})
	return p
}

func (p *QueryParser) Range(field string, values ...interface{}) *QueryParser {
	p.conditions = append(p.conditions, &condition{field: field, values: values, op: "range"})
	return p
}

func (p *QueryParser) InInt64(field string, values ...int64) *QueryParser {
	// convert to []interface{}{}
	var interfaceValues = []interface{}{}
	for _, v := range values {
		interfaceValues = append(interfaceValues, v)
	}
	p.In(field, interfaceValues...) // call
	return p
}

// add a `field`.in('v1', 'v2', ...)
func (p *QueryParser) In(field string, values ...interface{}) *QueryParser {
	p.conditions = append(p.conditions, &condition{
		field:  field,
		values: values,
		op:     "in",
	})
	return p
}

// TODO change this to OrderBy2
// func (p *QueryParser) OrderBy(orderby string) *QueryParser {
// 	p.orderby = orderby
// 	return p
// }

// order: asc | desc
func (p *QueryParser) OrderBy(orderby string, order string) *QueryParser {
	p.orderby = orderby
	p.order = order
	return p
}

func (p *QueryParser) DefaultOrderBy(orderby string, order string) *QueryParser {
	if p.orderby == "" {
		p.OrderBy(orderby, order)
	}
	return p
}

func (p *QueryParser) IsOrderBySet() bool {
	return p.orderby == ""
}

// e.g.: .Limit(1,10)
// e.g.: .Limit(1)
func (p *QueryParser) Limit(limit ...int) *QueryParser {
	if len(limit) >= 1 {
		p.limit = gxl.NewInt(limit[0])
	}
	if len(limit) >= 2 {
		p.n = gxl.NewInt(limit[1])
	}

	// p.debug_print_condition()

	return p
}

func (p *QueryParser) DefaultLimit(limit ...int) *QueryParser {
	if p.limit == nil {
		p.Limit(limit...)
	}
	return p
}

func (p *QueryParser) debug_print_condition() {
	if p.conditions != nil {
		for idx, con := range p.conditions {
			print(idx, " :: ", con.field, "\n")
		}
	}
}

// pin sql and cache them
func (p *QueryParser) Prepare() *QueryParser {
	if p.prepared {
		return p
	}

	e := p.e
	var sql bytes.Buffer
	switch p.operation {
	case "sql":
		// 1. The customized part, contains select .... from ... [join] , stoped before where
		sql.WriteString(p.fields[0])
		// 2. Where ... orderby ... limit...
		p.appendCommonStatementsAfterSelect(&sql)

	case "select":
		sql.WriteString("SELECT ")
		// fields
		var fields []string = e.Fields
		if p.useCustomerFields {
			fields = p.fields
		}
		sql.WriteString(fieldString(fields, e.Alias))

		// from <table>
		sql.WriteString(fromStatString(e.Table, e.Alias)) // from `table` as alias

		// after where in common select statements
		p.appendCommonStatementsAfterSelect(&sql)

	case "count":
		sql.WriteString("SELECT count(1)")
		sql.WriteString(fromStatString(e.Table, e.Alias)) // from `table` as alias

		// add where condition, default only support and
		if p.conditions != nil && len(p.conditions) > 0 {
			sql.WriteString(" WHERE ")
			p.values = appendWhereClouse(&sql, e.Alias, p.conditions...)
		}
	case "insert":
		// em.Insert().Exec(name, class, ...)
		sql.WriteString("insert into `")
		sql.WriteString(e.Table)
		sql.WriteString("` (")

		fields := e.CreateFields
		if p.useCustomerFields {
			fields = p.fields
		}
		sql.WriteString(fmt.Sprintf("`%v`", strings.Join(fields, "`,`")))
		sql.WriteString(" )")
		// values
		sql.WriteString(" values (")
		for i := 0; i < len(fields); i++ {
			if i > 0 {
				sql.WriteString(",")
			}
			sql.WriteString("?")
		}
		sql.WriteString(" )")

	case "update":
		// em.Update().Where("id", 5).Exec(name, class, ...)
		// em.Update().Exec(name, class, ..., id)
		sql.WriteString("update `")
		sql.WriteString(e.Table)
		sql.WriteString("` set ")

		fields := e.UpdateFields
		if p.useCustomerFields {
			fields = p.fields
		}
		for i := 0; i < len(fields); i++ {
			if i > 0 {
				sql.WriteString(",")
			}
			sql.WriteString(fmt.Sprintf("`%v`=?", fields[i]))
		}

		// where
		sql.WriteString(" WHERE ")
		if p.conditions == nil || len(p.conditions) == 0 {
			sql.WriteString(fmt.Sprintf(" `%v` = ?", e.PK))
		} else {
			p.values = appendWhereClouse(&sql, e.Alias, p.conditions...)
		}

	case "delete":
		// em.Delete().Where("id", 5).Exec()
		sql.WriteString("delete ")
		sql.WriteString(fromStatString(e.Table, "")) // delete don't need alias.

		// where
		sql.WriteString(" WHERE ")
		if p.conditions == nil || len(p.conditions) == 0 {
			sql.WriteString(fmt.Sprintf(" `%v` = ?", e.PK))
		} else {
			p.values = appendWhereClouse(&sql, e.Alias, p.conditions...)
			// // TODO ... to be condinued....
			// for i := 0; i < len(p.conditions); i = i + 2 {
			// 	k, v := p.where[i].(string), p.where[i+1]
			// 	sql.WriteString(fmt.Sprintf(" `%v` = ?", k))
			// 	p.values = append(p.values, v)
			// 	if i < len(p.where)-3 {
			// 		sql.WriteString(" and ")
			// 	}
			// }
		}

	}
	p.sql = sql.String()
	p.prepared = true
	return p
}

// sql statements after where in a common select sql. Including where, orderby limit.
func (p *QueryParser) appendCommonStatementsAfterSelect(sql *bytes.Buffer) {
	// Where condition, default only support and
	if p.conditions != nil && len(p.conditions) > 0 {
		sql.WriteString(" WHERE ")
		p.values = appendWhereClouse(sql, p.e.Alias, p.conditions...)
	}

	if p.orderby != "" {
		sql.WriteString(" order by ")
		if p.order == "" { // todo kill this
			sql.WriteString(p.orderby)
		} else { // leave this  // support OrderBy2
			if p.e.Alias != "" {
				sql.WriteString(p.e.Alias)
				sql.WriteRune('.')
			}
			sql.WriteString(p.orderby)
			sql.WriteRune(' ')
			sql.WriteString(p.order)
		}
	}

	if p.limit != nil {
		sql.WriteString(" limit ")
		sql.WriteString(p.limit.String())
		if p.n != nil {
			sql.WriteString(",")
			sql.WriteString(p.n.String())
		}
	}
}

// deprecated. why not this?
// param: use these value parameters to replace default value.
func (p *QueryParser) QueryOne(receiver func(*sql.Row) error) error {
	// query one will throw exceptions, so use query instead
	// TODO add limit support to QueryBuilder

	p.Prepare()

	// 1. get connection
	conn, err := Connect()
	if Err(err) {
		return err
	}
	defer conn.Close()

	// 2. prepare sql
	stmt, err := conn.Prepare(p.sql)
	if Err(err) {
		return err
	}
	defer stmt.Close()

	// 3. execute
	row := stmt.QueryRow(p.values...)
	if row != nil {
		err = receiver(row) // callbacks to receive values.
		if Err(err) {
			return err
		}
	}
	return nil
}

// query multi-results
// param: receiver is a callback function to process result set;
func (p *QueryParser) Query(receiver func(*sql.Rows) (bool, error)) error {
	p.Prepare()

	if sqllogger.Info() {
		sqllogger.Printf("------------- SQL ----------------\n  %s\n", p.sql)
		// fmt.Println("================= SQL Statement and it's values =================")
		// debuglog("Query", "\"%v\"", p.sql)
		sqllogger.Print("  Param: \n")
		for idx, v := range p.values {
			sqllogger.Printf("[%d: %v] ", idx, v)
		}
	}

	// 1. get connection
	conn, err := Connect()
	defer CloseConn(conn)
	if Err(err) {
		return err
	}

	// 2. prepare sql
	stmt, err := conn.Prepare(p.sql)
	if Err(err) {
		return err
	}
	defer CloseStmt(stmt)

	// 3. execute
	rows, err := stmt.Query(p.values...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		goon, err := receiver(rows) // callbacks to receive values.
		if Err(err) {
			return err
		}
		if !goon {
			break
		}
	}
	return nil
}

func (p *QueryParser) QueryInt() (int, error) {
	var count int
	if err := p.Query(
		func(rows *sql.Rows) (bool, error) {
			return false, rows.Scan(&count)
		},
	); err != nil {
		return -1, err
	}
	return count, nil
}

// exec command insert, update, delete
func (p *QueryParser) Exec(values ...interface{}) (sql.Result, error) {
	p.Prepare()

	debuglog("Exec", "\"%v\"", p.sql)

	var conn *sql.DB
	var stmt *sql.Stmt
	var err error
	if conn, err = Connect(); err != nil {
		return nil, err
	}
	defer conn.Close()

	if stmt, err = conn.Prepare(p.sql); err != nil {
		return nil, err
	}
	defer stmt.Close()

	// execute
	v := []interface{}{}
	v = append(v, values...)
	// for update command, use values as where condition.
	if p.values != nil && len(p.values) > 0 {
		v = append(v, p.values...)
	}

	debuglog("Exec", "with parameters %v", v)

	res, err := stmt.Exec(v...)
	if Err(err) {
		return nil, err
	}
	return res, nil
}

// ________________________________________________________________________________
var logEnabled = true

func debuglog(method string, format string, params ...interface{}) {
	if logEnabled {
		fmt.Printf("[DB.%v] %v\n",
			method,
			fmt.Sprintf(format, params...),
		)
	}
}

// helper methods
func fieldString(fields []string, alias string) string {
	if fields == nil || len(fields) == 0 {
		return "*"
	}
	if alias != "" {
		var buffer bytes.Buffer
		for idx, field := range fields {
			if idx > 0 {
				buffer.WriteString(", ")
			}
			buffer.WriteRune('\'')
			buffer.WriteString(alias)
			buffer.WriteRune('.')
			buffer.WriteString(field)
			buffer.WriteRune('\'')
		}
		return buffer.String()
	} else {
		return fmt.Sprintf("`%v`",
			strings.Join(fields, "`, `"),
		)
	}
}

func fromStatString(table string, alias string) string {
	var sql bytes.Buffer
	sql.WriteString(" FROM `")
	sql.WriteString(table)
	sql.WriteString("`")
	if alias != "" {
		sql.WriteString(" as ")
		sql.WriteString(alias)
	}
	return sql.String()
}

func appendWhereClouse(sql *bytes.Buffer, alias string, conditions ...*condition) []interface{} {
	values := []interface{}{}
	thefirst := true
	sql.WriteString(" ")
	for _, con := range conditions {
		lenvalue := len(con.values)
		switch con.op {
		case "and", "or":
			if !thefirst {
				if lenvalue > 1 {
					sql.WriteString(" and ")
				} else {
					sql.WriteString(" ")
					sql.WriteString(con.op)
					sql.WriteString(" ")
				}
			}
			if lenvalue == 1 {
				sql.WriteString(stringFieldEqualsQuestion(alias, con.field)) // p.`x`=?
				// sql.WriteString(fmt.Sprintf("`%v`=?", con.field))
			} else if lenvalue > 1 {
				sql.WriteString("(")
				for idx, _ := range con.values {
					sql.WriteString(stringFieldEqualsQuestion(alias, con.field))
					// sql.WriteString(fmt.Sprintf("`%v`=?", con.field))
					if idx < lenvalue-1 {
						sql.WriteString(" ")
						sql.WriteString(con.op)
						sql.WriteString(" ")
					}
				}
				sql.WriteString(")")
			}
			values = append(values, con.values...)

		case "in":
			if lenvalue == 0 {
				break
			}
			if !thefirst {
				sql.WriteString(" and ")
			}
			sql.WriteString(stringFieldWithAlisa(alias, con.field))
			sql.WriteString(" in (")
			for i := 0; i < lenvalue; i++ {
				if i > 0 {
					sql.WriteRune(',')
				}
				sql.WriteRune('?')
			}
			sql.WriteString(")")

			// fmt.Println("\n\n---------------------------------------------~~~~~~~~~~~~~~~~~--")
			// fmt.Println("<<<<<<|~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
			// fmt.Println("before ", values)
			values = append(values, con.values...)
			// fmt.Println("after ", values)
			// fmt.Println("values: ", con.values)

		case "like":
			if lenvalue == 0 {
				break
			}
			if !thefirst {
				sql.WriteString(" and ")
			}
			sql.WriteString(stringFieldWithAlisa(alias, con.field))
			sql.WriteString(" like ?")
			values = append(values, con.values...)

		case "range":
			if lenvalue == 0 || lenvalue > 2 {
				panic("Where clause must only have 1 or 2 values.")
			}
			if !thefirst {
				sql.WriteString(" and ")
			}
			sql.WriteString("(")
			sql.WriteString(fmt.Sprintf("%s>=?", stringFieldWithAlisa(alias, con.field)))
			if lenvalue > 1 {
				sql.WriteString(fmt.Sprintf(" and %s<?", stringFieldWithAlisa(alias, con.field)))
			}
			sql.WriteString(")")
			values = append(values, con.values...)

		case "sql": // condition's customized sql
			if !thefirst {
				sql.WriteString(" and ")
			}
			sql.WriteString(con.field)
			values = append(values, con.values...)
		}

		thefirst = false
	}
	sql.WriteString(" ")
	return values
}

// output c.`field` or `field`
func stringFieldWithAlisa(alias string, field string) string {
	if alias == "" {
		return fmt.Sprintf("`%s`", field)
	} else {
		return fmt.Sprintf("%s.`%s`", alias, field)
	}
}

func stringFieldEqualsQuestion(alias string, field string) string {
	// fmt.Sprintf("`%v`=?", con.field)
	if alias == "" {
		return fmt.Sprintf("`%v`=?", field)
	} else {
		return fmt.Sprintf("%s.`%v`=?", alias, field)
	}
}

// 先输入Where等条件，最后再从QueryParser来select或者count的方法。

func (qp *QueryParser) Select(fields ...string) *QueryParser {
	return qp.setSelectQueryParser("select", fields...)
}

func (qp *QueryParser) RawQuery(sql string) *QueryParser {
	return qp.setSelectQueryParser("sql", sql)
}

func (qp *QueryParser) Delete() *QueryParser {
	return qp.setSelectQueryParser("delete")
}

func (qp *QueryParser) Count() (int, error) {
	return qp.setSelectQueryParser("count").QueryInt()
}

func (qp *QueryParser) setSelectQueryParser(operation string, fields ...string) *QueryParser {
	qp.operation = operation
	qp.fields = fields
	if nil != fields && len(fields) > 0 {
		qp.useCustomerFields = true // only select clause use this
	}
	return qp
}
