package ecs

// EntityID is the identification of an entity in the world
type EntityID uint64

const (
	// bits used for the identifier part of the ID
	EntityIdentifierBitCount = 32
	// bits used to the generation part of the ID
	EntityGenerationBitCount = 24
	EntityGenerationShift    = EntityIdentifierBitCount
	EntityFlagsStartBit      = (EntityIdentifierBitCount + EntityGenerationBitCount)
	EntityIdentifierMask     = uint64(0x00000000ffffffff)
	EntityGenerationMask     = uint64(0x00ffffff00000000)
	EntityFlagsMask          = uint64(0xff00000000000000)
)

const (
	FlagEntityChildOf    = EntityID(1 << EntityFlagsStartBit)
	FlagEntityInstanceOf = EntityID(1 << (EntityFlagsStartBit + 1))
	FlagEntityDisabled   = EntityID(1 << (EntityFlagsStartBit + 2))
	FlagEntityComponent  = EntityID(1 << (EntityFlagsStartBit + 3))
	FlagEntitySingleton  = EntityID(1 << (EntityFlagsStartBit + 4))
)

// MakeEntity returns a new EntityID with id and generation
func MakeEntity(id, gen uint64) EntityID {
	entity := id | (gen << EntityGenerationShift & EntityGenerationMask)
	return EntityID(entity)
}

// MakeEntityWithFlags returns a new EntityID with id, generation and flags set
func MakeEntityWithFlags(id, gen uint64, flags EntityID) EntityID {
	return MakeEntity(id, gen) | flags
}

// UInt64 returns the full EntityID as uint64
func (e EntityID) UInt64() uint64 {
	return uint64(e)
}

// Gen returns the generation part of the EntityID
func (e EntityID) Gen() uint64 {
	return e.UInt64() & EntityGenerationMask >> EntityGenerationShift
}

// ID returns the identification part of the entity
func (e EntityID) ID() uint64 {
	return e.UInt64() & EntityIdentifierMask
}

// SetID returns a new EntityID with the new id and old generation and flags
func (e EntityID) SetID(id uint64) EntityID {
	return EntityID(e.UInt64()&^EntityIdentifierMask | (id & EntityIdentifierMask))
}

// Flags returns a new EntityID with only the flags set
func (e EntityID) Flags() EntityID {
	return EntityID(e.UInt64() & EntityFlagsMask)
}

func (e EntityID) IsInstance() bool {
	return e&FlagEntityInstanceOf != 0
}

func (e EntityID) IsChild() bool {
	return e&FlagEntityChildOf != 0
}

func (e EntityID) IsDisabled() bool {
	return e&FlagEntityDisabled != 0
}

func (e EntityID) IsComponent() bool {
	return e&FlagEntityComponent != 0
}

func (e EntityID) IsSingleton() bool {
	return e&FlagEntitySingleton != 0
}

func (e EntityID) ChildOf(enable bool) EntityID {
	if enable {
		return e | EntityID(FlagEntityChildOf)
	}
	return e & ^EntityID(FlagEntityChildOf)
}

func (e EntityID) InstanceOf(enable bool) EntityID {
	if enable {
		return e | EntityID(FlagEntityInstanceOf)
	}
	return e & ^EntityID(FlagEntityInstanceOf)
}

func (e EntityID) Disable() EntityID {
	return e | EntityID(FlagEntityDisabled)
}

func (e EntityID) Enable() EntityID {
	return e & ^EntityID(FlagEntityDisabled)
}

// WithoutFlags returns a new EntityID with only the id and generation
func (e EntityID) WithoutFlags() EntityID {
	return e & ^EntityID(EntityFlagsMask)
}

func (e EntityID) Component() EntityID {
	return MakeEntityWithFlags(e.ID(), 0, FlagEntityComponent)
}

func (e EntityID) Singleton() EntityID {
	return e.Component() | EntityID(FlagEntitySingleton)
}
