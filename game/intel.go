package game

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const Unexplored = -1
const Unowned = -1

type discover struct {
	game *Game
}

type discoverer interface {
	playerInfoDiscover(player *Player)
}

type Intel struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Dirty     bool      `json:"-" gorm:"-"`
	GameID    uint      `json:"gameId"`
	Name      string    `json:"name"`
	Num       int       `json:"num"`
	PlayerNum int       `json:"playerNum"`
	PlayerID  uint      `json:"-"`
	ReportAge int       `json:"reportAge"`
}

type MapObjectIntel struct {
	Intel
	Type     MapObjectType `json:"type"`
	Position Vector        `json:"position" gorm:"embedded"`
}

func (intel *Intel) String() string {
	return fmt.Sprintf("GameID: %5d, ID: %5d, Num: %3d %s", intel.GameID, intel.ID, intel.Num, intel.Name)
}

type PlanetIntel struct {
	MapObjectIntel
	Hab                  Hab         `json:"hab,omitempty" gorm:"embedded;embeddedPrefix:hab_"`
	MineralConcentration Mineral     `json:"mineralConcentration,omitempty" gorm:"embedded;embeddedPrefix:mineral_conc_"`
	Population           uint        `json:"population,omitempty"`
	Starbase             *FleetIntel `json:"starbase,omitempty"`
	Cargo                Cargo       `json:"cargo,omitempty" gorm:"embedded;embeddedPrefix:cargo_"`
	CargoDiscovered      bool        `json:"cargoDiscovered,omitempty"`
}

type ShipDesignIntel struct {
	Intel
	UUID          uuid.UUID        `json:"uuid,omitempty"`
	Name          string           `json:"name,omitempty"`
	Hull          string           `json:"hull,omitempty"`
	HullSetNumber int              `json:"hullSetNumber,omitempty"`
	Version       int              `json:"version,omitempty"`
	Armor         int              `json:"armor,omitempty"`
	Shields       int              `json:"shields,omitempty"`
	Slots         []ShipDesignSlot `json:"slots,omitempty" gorm:"serializer:json"`
}

type FleetIntel struct {
	MapObjectIntel
	PlanetIntelID   uint  `json:"-"` // for starbase fleets that are owned by a planet
	Cargo           Cargo `json:"cargo,omitempty" gorm:"embedded;embeddedPrefix:cargo_"`
	CargoDiscovered bool  `json:"cargoDiscovered,omitempty"`
}

type MineralPacketIntel struct {
	MapObjectIntel
	WarpFactor uint   `json:"warpFactor,omitempty"`
	Heading    Vector `json:"position" gorm:"embedded;embeddedPrefix:heading_"`
	Cargo      Cargo  `json:"cargo,omitempty" gorm:"embedded;embeddedPrefix:cargo_"`
}

type SalvageIntel struct {
	MapObjectIntel
	Cargo Cargo `json:"cargo,omitempty"`
}

type MineFieldIntel struct {
	MapObjectIntel
	NumMines uint          `json:"numMines,omitempty"`
	Type     MineFieldType `json:"type,omitempty"`
}

func (p *PlanetIntel) String() string {
	return fmt.Sprintf("Planet %s", &p.MapObjectIntel)
}

func (f *FleetIntel) String() string {
	return fmt.Sprintf("Player: %d, Fleet: %s", f.PlayerNum, f.Name)
}

func (d *ShipDesignIntel) String() string {
	return fmt.Sprintf("Player: %d, Fleet: %s", d.PlayerNum, d.Name)
}

// create a new FleetIntel object by key
func NewFleetIntel(playerNum int, name string) FleetIntel {
	return FleetIntel{
		MapObjectIntel: MapObjectIntel{
			Intel: Intel{
				Name:      name,
				PlayerNum: playerNum,
			},
		},
	}
}

// true if we haven't explored this planet
func (intel *PlanetIntel) Unexplored() bool {
	return intel.ReportAge == Unexplored
}

// true if we have explored this planet
func (intel *PlanetIntel) Explored() bool {
	return intel.ReportAge != Unexplored
}

// discover a planet and add it to the player's intel
func discoverPlanet(rules *Rules, player *Player, planet *Planet, penScanned bool) error {

	var intel *PlanetIntel
	planetIndex := planet.Num - 1

	if planetIndex < 0 || planetIndex >= len(player.PlanetIntels) {
		return fmt.Errorf("player %s cannot discover planet %s, planetIndex %d out of range", player, planet, planetIndex)
	}

	intel = &player.PlanetIntels[planetIndex]

	// if this intel is new, make sure it saves to the DB
	// once we create the object in the DB, it only gets saved to the DB
	// again if pen scanned
	if intel.PlayerID == 0 {
		intel.PlayerID = player.ID // this player owns this intel
		intel.Dirty = true
		intel.ReportAge = Unexplored
		intel.Type = MapObjectTypePlanet
		intel.PlayerNum = Unowned
	}

	// everyone knows these about planets
	intel.GameID = planet.GameID
	intel.Position = planet.Position
	intel.Name = planet.Name
	intel.Num = planet.Num

	ownedByPlayer := planet.PlayerNum != Unowned && player.Num == planet.PlayerNum

	if penScanned || ownedByPlayer {
		intel.Dirty = true // flag for update
		intel.PlayerNum = planet.PlayerNum

		// if we pen scanned the planet, we learn some things
		intel.Hab = planet.Hab
		intel.MineralConcentration = planet.MineralConcentration
		intel.ReportAge = 0

		// players know their planet pops, but other planets are slightly off
		if ownedByPlayer {
			intel.Population = uint(planet.Population())
		} else {
			var randomPopulationError = rules.Random.Float64()*(rules.PopulationScannerError-(-rules.PopulationScannerError)) - rules.PopulationScannerError
			intel.Population = uint(float64(planet.Population()) * (1 - randomPopulationError))
		}
	}
	return nil
}

// discover the cargo of a planet
func discoverPlanetCargo(player *Player, planet *Planet) error {

	var intel *PlanetIntel
	planetIndex := planet.Num - 1

	if planetIndex < 0 || planetIndex >= len(player.PlanetIntels) {
		return fmt.Errorf("player %s cannot discover planet %s, planetIndex %d out of range", player, planet, planetIndex)
	}

	intel = &player.PlanetIntels[planetIndex]

	intel.CargoDiscovered = true
	intel.Cargo = Cargo{
		Ironium:   planet.Cargo.Ironium,
		Boranium:  planet.Cargo.Boranium,
		Germanium: planet.Cargo.Germanium,
	}

	return nil

}

// discover a fleet and add it to the player's fleet intel
func discoverFleet(player *Player, fleet *Fleet) {
	intel := NewFleetIntel(fleet.PlayerNum, fleet.Name)

	intel.Name = fleet.Name
	intel.PlayerNum = fleet.PlayerNum
	intel.GameID = fleet.GameID
	intel.Position = fleet.Position

	player.FleetIntels = append(player.FleetIntels, intel)
	player.FleetIntelsByKey[intel.String()] = &intel
}

// discover cargo for an existing fleet
func discoverFleetCargo(player *Player, fleet *Fleet) {
	key := NewFleetIntel(fleet.PlayerNum, fleet.Name)

	existingIntel, found := player.FleetIntelsByKey[key.String()]
	if found {
		existingIntel.Cargo = fleet.Cargo
		existingIntel.CargoDiscovered = true
	}

}

// discover a player's design. This is a noop if we already know about
// the design and aren't discovering slots
func discoverDesign(player *Player, design *ShipDesign, discoverSlots bool) {
	intel, found := player.DesignIntelsByKey[design.UUID]
	if !found {
		// create a new intel for this design
		intel = &ShipDesignIntel{
			Intel: Intel{
				GameID:    design.GameID,
				Dirty:     true,
				Name:      design.Name,
				PlayerNum: design.PlayerNum,
			},
			UUID:          design.UUID,
			Hull:          design.Hull,
			HullSetNumber: design.HullSetNumber,
		}

		// save this new design to our intel
		player.DesignIntels = append(player.DesignIntels, *intel)
		intel = &player.DesignIntels[len(player.DesignIntels)-1]
		player.DesignIntelsByKey[intel.UUID] = intel
	}

	// discover slots if we haven't already and this scanner discovers them
	if discoverSlots && len(intel.Slots) == 0 {
		intel.Slots = make([]ShipDesignSlot, len(design.Slots))
		copy(intel.Slots, design.Slots)
		intel.Armor = design.Spec.Armor
		intel.Shields = design.Spec.Shield
	}
}

func (d discover) playerInfoDiscover(player *Player) {
	// d.game <- players to discover
	// discover info about other players
}
