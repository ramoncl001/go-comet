package comet

import "gorm.io/gorm"

type DatabaseContext struct {
	*gorm.DB
}

func NewDatabaseContext(dialector gorm.Dialector, args ...gorm.Option) *DatabaseContext {
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		panic(err)
	}

	return &DatabaseContext{
		db,
	}
}
