package common

// Add the string 'element' if not already in the array 'list'. Return the new array
func AddIfNotExists(element string, list []string) []string {
	n := len(list)
	for i := 0; i < n; i++ {
		if list[i] == element {
			return list
		}
	}
	list = list[0 : +n+1]
	list[n] = element
	return list
}

func AddIfNotExistsGeneric(element interface{}, list []interface{}) []interface{} {
	n := len(list)
	for i := 0; i < n; i++ {
		if list[i] == element {
			return list
		}
	}
	list = list[0 : +n+1]
	list[n] = element
	return list
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
