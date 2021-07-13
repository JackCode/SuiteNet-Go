package mysql

import (
	"database/sql"

	"github.com/jackcode/suitenet/pkg/models"
)

type LocationModel struct {
	DB *sql.DB
}

func (m *LocationModel) GetActiveLocations() ([]*models.Location, error) {
	stmt := `SELECT id, title, created FROM location WHERE is_active = TRUE`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	locations := []*models.Location{}

	for rows.Next() {
		location := &models.Location{
			CreatedBy: &models.SysUser{},
		}

		err = rows.Scan(&location.ID, &location.Title, &location.Created)
		if err != nil {
			return nil, err
		}

		locations = append(locations, location)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return locations, nil
}
