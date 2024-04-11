package core

var flags = map[Entity]map[int]bool{}

func SetFlag(entity Entity, field int, value bool) {
	if _, ok := flags[entity]; !ok {
		flags[entity] = map[int]bool{}
	}

	flags[entity][field] = value
}

func GetFlag(entity Entity, field int) bool {
	if _, ok := flags[entity]; ok {
		return flags[entity][field]
	}

	return false
}
