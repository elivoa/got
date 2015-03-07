package db

// Integrated method.
func (e *Entity) CountAll() (int, error) {
	return e.Count().Where().QueryInt()
}

// Integrated method.
func (e *Entity) CountBy(field string, value interface{}) (int, error) {
	return e.Count().Where(field, value).QueryInt()
}

// Integrated method.
func (e *Entity) DeleteByPK(pk interface{}) (int64, error) {
	res, err := e.Delete().Exec(pk)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
