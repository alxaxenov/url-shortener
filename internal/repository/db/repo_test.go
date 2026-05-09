package db

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/config/db/pg"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/model"
	repoModel "github.com/alxaxenov/url-shortener/tree/v2/internal/repository/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDBRepo_SetValue(t *testing.T) {
	sqlQuery := `INSERT INTO urls (short_url, original_url, created_at, user_id, active) VALUES($1, $2, $3, $4, $5) ON CONFLICT (original_url) DO UPDATE SET original_url = EXCLUDED.original_url RETURNING short_url`

	originURL := "http://example.com"
	userID := 42

	tests := []struct {
		name      string
		short     string
		mockSetup func(mock sqlmock.Sqlmock, short string)
		want      string
		expErr    func(*testing.T, error)
	}{
		{
			name:  "successful insert",
			short: "short_1",
			mockSetup: func(mock sqlmock.Sqlmock, short string) {
				mock.ExpectQuery(sqlQuery).
					WithArgs(short, originURL, sqlmock.AnyArg(), userID, true).
					WillReturnRows(sqlmock.NewRows([]string{"short_url"}).AddRow("short_1"))
			},
			want: "short_1",
			expErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:  "insert returns existing short_url on conflict (upsert)",
			short: "short_2",
			mockSetup: func(mock sqlmock.Sqlmock, short string) {
				mock.ExpectQuery(sqlQuery).
					WithArgs(short, originURL, sqlmock.AnyArg(), userID, true).
					WillReturnRows(sqlmock.NewRows([]string{"short_url"}).AddRow("existing456"))
			},
			want: "existing456",
			expErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:  "database error on query",
			short: "short_3",
			mockSetup: func(mock sqlmock.Sqlmock, short string) {
				mock.ExpectQuery(sqlQuery).
					WithArgs(short, originURL, sqlmock.AnyArg(), userID, true).
					WillReturnError(errors.New("insert_test_error"))
			},
			want: "",
			expErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				expectedPrefix := "SetValue failed to insert url"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer db.Close()

			connector := &pg.ConnectorPG{}
			connector.DB = db
			repo := &DBRepo{Connector: connector}
			tt.mockSetup(mock, tt.short)

			got, err := repo.SetValue(context.Background(), tt.short, originURL, userID)

			tt.expErr(t, err)
			assert.Equalf(t, tt.want, got, "SetValue - %s, unexpected result", tt.name)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDBRepo_GetValue(t *testing.T) {
	sqlQuery := `SELECT original_url, active FROM urls WHERE short_url=$1`

	tests := []struct {
		name       string
		short      string
		wantOrigin string
		wantActive bool
		mockSetup  func(mock sqlmock.Sqlmock, short string)
		expErr     func(*testing.T, error)
	}{
		{
			name:  "successful get active",
			short: "short_1",
			mockSetup: func(mock sqlmock.Sqlmock, short string) {
				mock.ExpectQuery(sqlQuery).
					WithArgs(short).
					WillReturnRows(sqlmock.NewRows([]string{"short_url", "active"}).AddRow("http://example.com", true))
			},
			wantOrigin: "http://example.com",
			wantActive: true,
			expErr:     func(t *testing.T, err error) { assert.NoError(t, err) },
		},
		{
			name:  "successful get not active",
			short: "short_2",
			mockSetup: func(mock sqlmock.Sqlmock, short string) {
				mock.ExpectQuery(sqlQuery).
					WithArgs(short).
					WillReturnRows(sqlmock.NewRows([]string{"short_url", "active"}).AddRow("http://example2.com", false))
			},
			wantOrigin: "http://example2.com",
			wantActive: false,
			expErr:     func(t *testing.T, err error) { assert.NoError(t, err) },
		},
		{
			name:  "get error on query",
			short: "short_2",
			mockSetup: func(mock sqlmock.Sqlmock, short string) {
				mock.ExpectQuery(sqlQuery).
					WithArgs(short).
					WillReturnError(errors.New("get_test_error"))
			},
			wantOrigin: "",
			wantActive: false,
			expErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				expectedPrefix := "GetValue failed to fetch url:"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer db.Close()

			connector := &pg.ConnectorPG{}
			connector.DB = db
			repo := &DBRepo{Connector: connector}
			tt.mockSetup(mock, tt.short)

			resOrigin, resActive, err := repo.GetValue(context.Background(), tt.short)

			tt.expErr(t, err)
			assert.Equalf(t, tt.wantOrigin, resOrigin, "GetValue.wantOrigin - %s, unexpected result", tt.name)
			assert.Equalf(t, tt.wantActive, resActive, "GetValue.wantActive - %s, unexpected result", tt.name)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDBRepo_CreateUser(t *testing.T) {
	sqlQuery := `INSERT INTO users (created_at) VALUES($1) RETURNING id`

	tests := []struct {
		name      string
		want      int
		mockSetup func(mock sqlmock.Sqlmock)
		expErr    func(*testing.T, error)
	}{
		{
			name: "successful create user",
			want: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(sqlQuery).
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
			},
			expErr: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "error query create user",
			want: 0,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(sqlQuery).
					WithArgs(sqlmock.AnyArg()).
					WillReturnError(errors.New("create_user_test_error"))
			},
			expErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				expectedPrefix := "CreateUser failed to create:"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer db.Close()

			connector := &pg.ConnectorPG{}
			connector.DB = db
			repo := &DBRepo{Connector: connector}
			tt.mockSetup(mock)

			res, err := repo.CreateUser(context.Background())

			tt.expErr(t, err)
			assert.Equalf(t, tt.want, res, "CreateUser - %s, unexpected result", tt.name)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// Кастомный конвертер, превращающий []string в PostgreSQL-массив (строку вида {a,b,c})
type pgxArrayConverter struct{}

func (c pgxArrayConverter) ConvertValue(v interface{}) (driver.Value, error) {
	switch val := v.(type) {
	case []string:
		// Формат PostgreSQL массива: {элемент1,элемент2}
		return "{" + strings.Join(val, ",") + "}", nil
	default:
		return driver.DefaultParameterConverter.ConvertValue(v)
	}
}

func TestDBRepo_DeleteURLs(t *testing.T) {
	sqlQuery := `UPDATE urls SET active = false WHERE user_id = $1 AND short_url = ANY($2) AND active = true`
	ctx := context.Background()
	userID := 42

	tests := []struct {
		name      string
		deleteReq *model.DeleteRequest
		mockSetup func(mock sqlmock.Sqlmock)
		want      int
		expErr    func(*testing.T, error)
	}{
		{
			name: "successful update",
			deleteReq: &model.DeleteRequest{
				UserID: userID,
				URLs:   []string{"a", "b", "c", "a"},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(sqlQuery).
					WithArgs(userID, []string{"a", "b", "c"}).
					WillReturnResult(sqlmock.NewResult(0, 3))
			},
			want: 3,
			expErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "no rows affected",
			deleteReq: &model.DeleteRequest{
				UserID: userID,
				URLs:   []string{"nonexistent"},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(sqlQuery).
					WithArgs(userID, []string{"nonexistent"}).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			want: 0,
			expErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "database error on ExecContext",
			deleteReq: &model.DeleteRequest{
				UserID: userID,
				URLs:   []string{"a", "b", "c"},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(sqlQuery).
					WithArgs(userID, []string{"a", "b", "c"}).
					WillReturnError(errors.New("connection lost"))
			},
			want: 0,
			expErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				expectedPrefix := "DeleteURLs error"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name: "RowsAffected returns error",
			deleteReq: &model.DeleteRequest{
				UserID: userID,
				URLs:   []string{"a", "b", "c"},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(sqlQuery).
					WithArgs(userID, []string{"a", "b", "c"}).
					WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))
			},
			want: 0,
			expErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(
				sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual),
				sqlmock.ValueConverterOption(pgxArrayConverter{}),
			)
			require.NoError(t, err)
			defer db.Close()

			connector := &pg.ConnectorPG{}
			connector.DB = db
			repo := &DBRepo{Connector: connector}
			tt.mockSetup(mock)

			gotAffected, err := repo.DeleteURLs(ctx, tt.deleteReq)

			tt.expErr(t, err)
			assert.Equalf(t, tt.want, gotAffected, "DeleteURLs - %s, unexpected result", tt.name)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDBRepo_UserURLs(t *testing.T) {
	sqlQuery := `SELECT short_url, original_url FROM urls WHERE user_id = $1 and active = true`

	tests := []struct {
		name      string
		want      []model.UserURLs
		mockSetup func(mock sqlmock.Sqlmock)
		expErr    func(*testing.T, error)
	}{
		{
			name: "query error",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(sqlQuery).
					WithArgs(13).
					WillReturnError(errors.New("query error"))
			},
			want: nil,
			expErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				expectedPrefix := "UserURLs failed to fetch urls:"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name: "scan error",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(sqlQuery).
					WithArgs(13).
					WillReturnRows(sqlmock.NewRows([]string{"short_url"}).AddRow(""))
			},
			want: nil,
			expErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				expectedPrefix := "UserURLs failed to scan url:"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name: "iterate error",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(sqlQuery).
					WithArgs(13).
					WillReturnRows(
						sqlmock.NewRows([]string{"short_url", "original_url"}).
							AddRow("a", "http://a.com").
							AddRow("b", "http://b.com").
							RowError(1, errors.New("failed to fetch row from database")),
					)
			},
			want: nil,
			expErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				expectedPrefix := "UserURLs rows.Err() failed:"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name: "iterate error",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(sqlQuery).
					WithArgs(13).
					WillReturnRows(
						sqlmock.NewRows([]string{"short_url", "original_url"}).
							AddRow("a", "http://a.com").
							AddRow("b", "http://b.com").
							RowError(1, errors.New("failed to fetch row from database")),
					)
			},
			want: nil,
			expErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				expectedPrefix := "UserURLs rows.Err() failed:"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		}, {
			name: "success",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(sqlQuery).
					WithArgs(13).
					WillReturnRows(
						sqlmock.NewRows([]string{"short_url", "original_url"}).
							AddRow("a", "http://a.com").
							AddRow("b", "http://b.com"),
					)
			},
			want: []model.UserURLs{{"a", "http://a.com"}, {"b", "http://b.com"}},
			expErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer db.Close()

			connector := &pg.ConnectorPG{}
			connector.DB = db
			repo := &DBRepo{Connector: connector}
			tt.mockSetup(mock)

			got, err := repo.UserURLs(context.Background(), 13)
			tt.expErr(t, err)
			assert.Equalf(t, tt.want, got, "UserURLs - %s, unexpected result", tt.name)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDBRepo_SaveBatch(t *testing.T) {
	tests := []struct {
		name      string
		batches   []repoModel.UploadBatch
		mockSetup func(mock sqlmock.Sqlmock)
		expErr    func(*testing.T, error)
	}{
		{
			name:      "empty batch",
			batches:   []repoModel.UploadBatch{},
			mockSetup: func(mock sqlmock.Sqlmock) {},
			expErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				expectedPrefix := "SaveBatch no batches to upload"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name:    "exec err",
			batches: []repoModel.UploadBatch{{"a", "http://a.com"}},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO urls (short_url, original_url, created_at, user_id, active) VALUES ($1,$2,$3,$4,$5)`).
					WithArgs("a", "http://a.com", sqlmock.AnyArg(), 42, true).
					WillReturnError(errors.New("exec failed"))
			},
			expErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				expectedPrefix := "SaveBatch failed to insert url:"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name:    "success",
			batches: []repoModel.UploadBatch{{"a", "http://a.com"}, {"b", "http://b.com"}},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO urls (short_url, original_url, created_at, user_id, active) VALUES ($1,$2,$3,$4,$5),($6,$7,$8,$9,$10)`).
					WithArgs("a", "http://a.com", sqlmock.AnyArg(), 42, true, "b", "http://b.com", sqlmock.AnyArg(), 42, true).
					WillReturnResult(sqlmock.NewResult(2, 2))
			},
			expErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer db.Close()

			connector := &pg.ConnectorPG{}
			connector.DB = db
			repo := &DBRepo{Connector: connector}
			tt.mockSetup(mock)

			err = repo.SaveBatch(context.Background(), tt.batches, 42)
			tt.expErr(t, err)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
