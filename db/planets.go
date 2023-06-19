package db

import (
	"errors"

	"github.com/sirgwain/craig-stars/game"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (db *DB) FindPlanetById(id uint) (*game.Planet, error) {
	planet := game.Planet{}
	if err := db.sqlDB.Preload(clause.Associations).First(&planet, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return &planet, nil
}

func (db *DB) SavePlanet(planet *game.Planet) error {
	if err := db.sqlDB.Save(planet).Error; err != nil {
		return err
	}

	err := db.sqlDB.Model(planet).Association("ProductionQueue").Replace(planet.ProductionQueue)

	return err
}
