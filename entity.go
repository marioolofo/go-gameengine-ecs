package ecs

type EntityID uint64

const (
	EntityIdentifierBitCount = 32
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

type EntityIDSlice []EntityID

func (e EntityIDSlice) Len() int           { return len(e) }
func (e EntityIDSlice) Less(i, j int) bool { return e[i] < e[j] }
func (e EntityIDSlice) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }

func MakeEntity(id, gen uint64) EntityID {
	entity := id | (gen << EntityGenerationShift & EntityGenerationMask)
	return EntityID(entity)
}

func MakeEntityWithFlags(id, gen uint64, flags EntityID) EntityID {
	return MakeEntity(id, gen) | flags
}

func (e EntityID) UInt64() uint64 {
	return uint64(e)
}

func (e EntityID) Gen() uint64 {
	return e.UInt64() & EntityGenerationMask >> EntityGenerationShift
}

func (e EntityID) ID() uint64 {
	return e.UInt64() & EntityIdentifierMask
}

func (e EntityID) SetID(id uint64) EntityID {
	return EntityID(e.UInt64()&^EntityIdentifierMask | (id & EntityIdentifierMask))
}

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

func (e EntityID) WithoutFlags() EntityID {
	return e & ^EntityID(EntityFlagsMask)
}

func (e EntityID) Component() EntityID {
	return MakeEntityWithFlags(e.ID(), 0, FlagEntityComponent)
}

func (e EntityID) Singleton() EntityID {
	return e.Component() | EntityID(FlagEntitySingleton)
}

func MakeEntityFindFn(entities []EntityID, find EntityID) (func(int) int) {
	return func(pos int) int {
		if entities[pos] == find {
			return 0
		}
		if entities[pos] < find {
			return 1
		}
		return -1
	}
}
