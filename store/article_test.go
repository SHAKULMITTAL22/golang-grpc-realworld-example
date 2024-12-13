package store

import (
	"fmt"
	"log"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
)

type ArticleStore struct {
	db *gorm.DB
}
type DB struct {
	sync.RWMutex
	Value			interface{}
	Error			error
	RowsAffected		int64
	db			SQLCommon
	blockGlobalUpdate	bool
	logMode			logModeValue
	logger			logger
	search			*search
	values			sync.Map
	parent			*DB
	callbacks		*Callback
	dialect			Dialect
	singularTable		bool
	nowFuncOverride		func() time.Time
}// single db
// function to be used to override the creating of a new timestamp


/*
 */
func (s *ArticleStore) GetByID(id uint) (*model.Article, error) {
	var m model.Article
	err := s.db.Preload("Tags").Preload("Author").Find(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

/*
 */
func TestArticleStoreGetByID(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("failed to open gorm DB, %v", err)
	}

	store := &ArticleStore{db: gormDB}

	type testCase struct {
		name     string
		id       uint
		mockFunc func()
		expected *model.Article
		err      error
	}

	tests := []testCase{
		{
			name: "Successful Retrieval of an Article by ID",
			id:   1,
			mockFunc: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "content"}).
					AddRow(1, "Test Title", "Test Content")
				mock.ExpectQuery("^SELECT (.+) FROM \"articles\" WHERE (.+)$").
					WithArgs(1).
					WillReturnRows(rows)

				mock.ExpectQuery("^SELECT (.+) FROM \"tags\" WHERE (.+)$").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Go"))
				mock.ExpectQuery("^SELECT (.+) FROM \"authors\" WHERE (.+)$").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Author Name"))
			},
			expected: &model.Article{
				Model:   gorm.Model{ID: 1},
				Title:   "Test Title",
				Content: "Test Content",
				Tags:    []model.Tag{{Model: gorm.Model{ID: 1}, Name: "Go"}},
				Author:  model.Author{Model: gorm.Model{ID: 1}, Name: "Author Name"},
			},
			err: nil,
		},
		{
			name: "Article Not Found",
			id:   2,
			mockFunc: func() {
				mock.ExpectQuery("^SELECT (.+) FROM \"articles\" WHERE (.+)$").
					WithArgs(2).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expected: nil,
			err:      gorm.ErrRecordNotFound,
		},
		{
			name: "Database Error During Retrieval",
			id:   3,
			mockFunc: func() {
				mock.ExpectQuery("^SELECT (.+) FROM \"articles\" WHERE (.+)$").
					WithArgs(3).
					WillReturnError(fmt.Errorf("database error"))
			},
			expected: nil,
			err:      fmt.Errorf("database error"),
		},
		{
			name: "Edge Case with ID Zero",
			id:   0,
			mockFunc: func() {
				mock.ExpectQuery("^SELECT (.+) FROM \"articles\" WHERE (.+)$").
					WithArgs(0).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expected: nil,
			err:      gorm.ErrRecordNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()

			article, err := store.GetByID(tc.id)

			assert.Equal(t, tc.expected, article)
			assert.Equal(t, tc.err, err)

			if err := mock.ExpectationsWereMet(); err != nil {
				log.Fatalf("there were unfulfilled expectations: %s", err)
			}

			t.Logf("Test %s: Success", tc.name)
		})
	}
}

/*
 */
func (s *UserStore) GetByID(id uint) (*model.User, error) {
	var m model.User
	if err := s.db.Find(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

