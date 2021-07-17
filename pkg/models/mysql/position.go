package mysql

import (
	"database/sql"

	"github.com/jackcode/suitenet/pkg/models"
)

type PositionModel struct {
	DB *sql.DB
}

func (m *PositionModel) GetActivePositions() ([]*models.Position, error) {
	stmt := `SELECT id, title, created FROM position WHERE is_active = TRUE AND title != "System Administrator"`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	positions := []*models.Position{}

	for rows.Next() {
		position := &models.Position{
			CreatedBy: &models.SysUser{},
		}

		err = rows.Scan(&position.ID, &position.Title, &position.Created)
		if err != nil {
			return nil, err
		}

		positions = append(positions, position)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return positions, nil
}
