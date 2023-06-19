package game

type Cargo struct {
	Ironium   int `json:"ironium,omitempty"`
	Boranium  int `json:"boranium,omitempty"`
	Germanium int `json:"germanium,omitempty"`
	Colonists int `json:"colonists,omitempty"`
}

func (c Cargo) Add(other Cargo) Cargo {
	return Cargo{
		Ironium:   c.Ironium + other.Ironium,
		Boranium:  c.Boranium + other.Boranium,
		Germanium: c.Germanium + other.Germanium,
		Colonists: c.Colonists + other.Colonists,
	}
}

func (c Cargo) Subtract(other Cargo) Cargo {
	return Cargo{
		Ironium:   c.Ironium - other.Ironium,
		Boranium:  c.Boranium - other.Boranium,
		Germanium: c.Germanium - other.Germanium,
		Colonists: c.Colonists - other.Colonists,
	}
}

func (c Cargo) AddMineral(other Mineral) Cargo {
	return Cargo{
		Ironium:   c.Ironium + other.Ironium,
		Boranium:  c.Boranium + other.Boranium,
		Germanium: c.Germanium + other.Germanium,
		Colonists: c.Colonists,
	}
}

func (c Cargo) ToMineral() Mineral {
	return Mineral{
		Ironium:   c.Ironium,
		Boranium:  c.Boranium,
		Germanium: c.Germanium,
	}
}

func (c Cargo) Total() int {
	return c.Ironium + c.Boranium + c.Germanium + c.Colonists
}

// return true if this cargo can have transferAmount taken from it
func (c Cargo) CanTransfer(transferAmount Cargo) bool {
	return (c.Ironium >= transferAmount.Ironium &&
		c.Boranium >= transferAmount.Boranium &&
		c.Germanium >= transferAmount.Germanium &&
		c.Colonists >= transferAmount.Colonists)

}
