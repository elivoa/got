package db

// Integrated method.
func (e *Entity) CountAll() (int, error) {
	return e.Count().Where().QueryInt()
}

// Integrated method.
func (e *Entity) CountBy(field string, value interface{}) (int, error) {
	return e.Count().Where(field, value).QueryInt()
}
