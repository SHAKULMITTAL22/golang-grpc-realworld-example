package store

import (
	"errors"
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
)





type ArticleStore struct {
	db gormDB
}
type gormDB interface {
	Table(name string) *gorm.DB
	Where(query interface{}, args ...interface{}) *gorm.DB
	Count(value interface{}) *gorm.DB
}
type mockDB struct {
	countResult int
	countError  error
}


/*
ROOST_METHOD_HASH=IsFavorited_799826fee5
ROOST_METHOD_SIG_HASH=IsFavorited_f6d5e67492

FUNCTION_DEF=func (s *ArticleStore) IsFavorited(a *model.Article, u *model.User) (bool, error) // IsFavorited returns whether the article is favorited by the user


*/
func (s *ArticleStore) IsFavorited(a *model.Article, u *model.User) (bool, error) {
	if a == nil || u == nil {
		return false, nil
	}
	var count int
	err := s.db.Table("favorite_articles").Where("article_id = ? AND user_id = ?", a.ID, u.ID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func TestArticleStoreIsFavorited(t *testing.T) {
	tests := []struct {
		name            string
		article         *model.Article
		user            *model.User
		mockCountResult int
		mockCountError  error
		want            bool
		wantErr         bool
	}{
		{
			name:            "Article is favorited by the user",
			article:         &model.Article{Model: gorm.Model{ID: 1}},
			user:            &model.User{Model: gorm.Model{ID: 1}},
			mockCountResult: 1,
			mockCountError:  nil,
			want:            true,
			wantErr:         false,
		},
		{
			name:            "Article is not favorited by the user",
			article:         &model.Article{Model: gorm.Model{ID: 1}},
			user:            &model.User{Model: gorm.Model{ID: 1}},
			mockCountResult: 0,
			mockCountError:  nil,
			want:            false,
			wantErr:         false,
		},
		{
			name:            "Nil Article parameter",
			article:         nil,
			user:            &model.User{Model: gorm.Model{ID: 1}},
			mockCountResult: 0,
			mockCountError:  nil,
			want:            false,
			wantErr:         false,
		},
		{
			name:            "Nil User parameter",
			article:         &model.Article{Model: gorm.Model{ID: 1}},
			user:            nil,
			mockCountResult: 0,
			mockCountError:  nil,
			want:            false,
			wantErr:         false,
		},
		{
			name:            "Database error",
			article:         &model.Article{Model: gorm.Model{ID: 1}},
			user:            &model.User{Model: gorm.Model{ID: 1}},
			mockCountResult: 0,
			mockCountError:  errors.New("database error"),
			want:            false,
			wantErr:         true,
		},
		{
			name:            "Multiple favorites for the same article and user",
			article:         &model.Article{Model: gorm.Model{ID: 1}},
			user:            &model.User{Model: gorm.Model{ID: 1}},
			mockCountResult: 2,
			mockCountError:  nil,
			want:            true,
			wantErr:         false,
		},
		{
			name:            "Zero count but no error from database",
			article:         &model.Article{Model: gorm.Model{ID: 1}},
			user:            &model.User{Model: gorm.Model{ID: 1}},
			mockCountResult: 0,
			mockCountError:  nil,
			want:            false,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &mockDB{
				countResult: tt.mockCountResult,
				countError:  tt.mockCountError,
			}

			s := &ArticleStore{
				db: mockDB,
			}

			got, err := s.IsFavorited(tt.article, tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleStore.IsFavorited() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ArticleStore.IsFavorited() = %v, want %v", got, tt.want)
			}
		})
	}
}

