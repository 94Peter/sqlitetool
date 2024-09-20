package sqlitetool

import (
	"bytes"
	"errors"
	"os"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SqlDI interface {
	NewSqlMgr() SqlMgr
}

type SqlMgr interface {
	NewDb(name string) (*DB, error)
	IsExist(name string) bool
}

type SqlConf struct {
	Path string `yaml:"path"`

	db      *DB
	newOnce sync.Once
	dbname  string
}

func (sc *SqlConf) NewSqlMgr() SqlMgr {
	return sc
}

func (sf *SqlConf) getFileName(name string) string {
	var b bytes.Buffer
	b.WriteString(sf.Path)
	b.WriteString(name)
	b.WriteString(".db")
	return b.String()
}
func (sf *SqlConf) NewDb(name string) (*DB, error) {
	sf.newOnce.Do(func() {
		db, err := gorm.Open(sqlite.Open(sf.getFileName(name)), &gorm.Config{})
		if err != nil {
			panic(err)
		}
		sf.db = &DB{DB: db}
		sf.dbname = name
	})
	if sf.dbname != name {
		return nil, errors.New("dbname not match, one application use one db")
	}
	return sf.db, nil
}

func (sf *SqlConf) IsExist(name string) bool {
	filename := sf.getFileName(name)
	return fileExists(filename)
}

type DB struct {
	*gorm.DB
}

func (s *DB) InitSqlDao(daos ...SqlDao) error {
	var err error
	for _, d := range daos {
		err = s.AutoMigrate(d)
		if err != nil {
			return err
		}
		if err = d.Init(s.DB); err != nil {
			return err
		}
	}
	return nil
}

type SqlDao interface {
	Init(mgr *gorm.DB) error
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
