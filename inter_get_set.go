package jsoninterpreter

import (
	"fmt"
	"strconv"
	"errors"
)

// Only this function commented, other Get() and Set() functions based on same logic. 
func Get(json []byte, path ... string) ([]byte, error){
	// path null.
	if len(path) == 0 {
		return nil, errors.New("Error: Path can not be null.")
	}
	// main offset track of this search.
	offset := 0
	currentPath := path[0]
	// important chars for json interpretation.
	// 34 = "
	// 44 = ,
	// 58 = :
	// 91 = [
	// 93 = ]
	// 123 = {
	// 125 = }
	chars := []byte{34, 44, 58, 91, 93, 123, 125}
	// creating a bool array fill with false
	isJsonChar := make([]bool, 256)
	// only interested chars is true
	for _,v := range chars {
		isJsonChar[v] = true
	}
	// trim spaces of start
	for space(json[offset]) {
		offset++
	}
	// braceType determine whether or not search will be a json search or array search
	braceType := json[offset]
	// main iteration off all bytes.
	for k := 0 ; k < len(path) ; k ++ {
		// 91 = [, beginning of an array search
		if braceType == 91 {
			// ARRAY SEACH SCOPE
			// path value cast to integer for determine index.
			arrayIndex, err := strconv.Atoi(currentPath)
			if err != nil {
				// braceType and current path type is conflicts.
				return nil, errors.New("Error: Index Expected, got string.")
			}
			// main done flag
			done := false
			// zeroth index search.
			if arrayIndex == 0 {
				// Increment offset for not catch current brace.
				offset++
				// Inner iteration for brace search.
				for i := offset; i < len(json) ; i ++ {
					// curr is current byte of reading.
					curr := json[i]
					// Open curly brace
					if curr == 123 {
						// change brace type of next search.
						braceType = curr
						if k != len(path) - 1{
							// If its not last path than change currentPath to next path.
							currentPath = path[k + 1]
						}
						// Assign offset to brace index.
						offset = i
						// Break the array search scope.
						done = true
						break
					}
					// Open square brace
					if curr == 91 {
						// change brace type of next search.
						braceType = curr
						if k != len(path) - 1{
							// If its not last path than change currentPath to next path.
							currentPath = path[k + 1]
						}
						// Searching for zeroth index is conflicts with searching zeroth array or arrays zeroth element.
						offset = i + 1
						// Break the array search scope.
						done = true
						break
					}
					// Doesn't have to always find a brace. It can be a value.
					if !space(curr){
						done = true
						break
					}
				}
			}else{
				// Brace level every brace increments the level
				level := 0
				// main in quote flag for determine what is in quote and what is not
				inQuote := false
				// index found flag.
				found := false
				// Index count of current element.
				indexCount := 0
				// Not interested with column char in this search
				isJsonChar[58] = false
				for i := offset ; i < len(json) ; i ++ {
					// curr is current byte of reading.
					curr := json[i]
					// Just interested with json chars. Other wise continue.
					if !isJsonChar[curr]{
						continue
					}
					// If current byte is quote
					if curr == 34 {
						// check before char it might be escape char.
						if json[i - 1] == 92 {
							continue
						}
						// Change inQuote flag to opposite.
						inQuote = !inQuote
						continue
					}
					if inQuote {
						continue
					}else{
						// Open braces
						if curr == 91 || curr == 123{
							// if found before done with this search
							// break array search scope
							if found {
								level++
								braceType = curr
								currentPath = path[k + 1]
								found = false
								done = true
								break
							}
							level++
							continue
						}
						if curr == 93 || curr == 125 {
							level--
							// if level is less than 1 it mean index not in this array. 
							if level < 1 {
								done = false
							}
							continue
						}
						// Not found before
						if !found {
							// same level with path
							if level == 1 {
								// current byte is comma
								if curr == 44 {
									// Inc index
									indexCount++
									if indexCount == arrayIndex {
										offset = i + 1
										if k == len(path) - 1{
											// last path found, break
											done = true
											break
										}
										// not last path keep going. For find next brace Type.
										found = true
										continue
									}
									continue
								}
								continue
							}
							continue
						}
						continue
					}
				}
				// Check true for column char again for keep same with first declaration.
				isJsonChar[58] = true
			}
			if !done {
				return nil, errors.New("Error: Index out of range")
			}
		}else{
			// KEY SEACH SCOPE
			// main in quote flag for determine what is in quote and what is not.
			inQuote := false
			// Key found flag.
			found := false
			// Key start index.
			start := 0
			// Key end index.
			end := 0
			// Current level.
			level := k
			// Not interested with comma in this search
			isJsonChar[44] = false
			for i := offset ; i < len(json) ; i ++ {
				// curr is current byte of reading.
				curr := json[i]
				// Just interested with json chars. Other wise continue.
				if !isJsonChar[curr]{
					continue
				}
				// If current byte is quote
				if curr == 34 {
					// change inQuote flag to opposite.
					inQuote = !inQuote
					// If key found no need to determine start and end points.
					if found {
						continue
					}
					// If level not same as path level no need to determine start and end points.
					if level != k + 1 {
						continue
					}
					// If starting new quote that means key starts here
					if inQuote {
						start = i + 1
						continue
					}
					// if quote ends that means key ends here
					end = i
					continue
				}
				if inQuote {
					continue
				}else{
					// open square brace
					if curr == 91 {
						// if found and new brace is square brace than 
						// next search is array search break loop and
						// update the current path 
						if found {
							braceType = curr
							currentPath = path[k + 1]
							break
						}
						level++
						continue
					}
					if curr == 123 {
						// if found and new brace is curly brace than 
						// next search is key search continue with this loop and
						// update the current path 
						// close found flag for next search.
						if found {
							k++
							level++
							currentPath = path[k]
							found = false
							continue
						}
						level++
						continue
					}
					// Close brace
					if curr == 93 || curr == 125 {
						level--
						continue
					}
					// same level with path
					if level == k + 1 {
						// column
						if curr == 58 {
							// compare key to current path 
							if compare(json, start, end, currentPath) {
								offset = i + 1
								found = true
								// if it is the last path element break
								// and include comma element to json chars.
								if k == len(path) - 1{
									isJsonChar[44] = true
									break
								}else{
									continue
								}
							}
							// Include comma element to json chars for jump function
							isJsonChar[44] = true
							// exclude column element to json chars for jump function
							isJsonChar[58] = false
							// jump function start :{} -> ,
							// it is fast travel from column to comma
							// first we need keys 
							// for this purpose skipping values. 
							// Only need value if key is correct
							for j := i ;  j < len(json) ; j ++ {
								// curr is current byte of reading.
								curr := json[j]
								// Just interested with json chars. Other wise continue.
								if !isJsonChar[curr]{
									continue
								}
								// Quote
								if curr == 34 {
									// check before char it might be escape char.
									if json[j - 1] == 92 {
										continue
									}
									// Change inQuote flag to opposite.
									inQuote = !inQuote
									continue
								}
								if inQuote {
									continue
								}else{
									// This brace conditions for level trace
									// it is necessary to keep level value correct
									if curr == 91 || curr == 123 {
										level++
										continue
									}
									if curr == 93 || curr == 125 {
										level--
										continue
									}
									// comma
									if curr == 44 {
										// level same with path
										if level == k + 1 {
											// jump i to j
											i = j
											break
										}
										continue
									}
									continue
								}

							}
							// exclude comma element to json chars, jump func is ending.
							isJsonChar[44] = false
							// Include column element to json chars, jump func is ending.
							isJsonChar[58] = true
							continue
						}
						continue
					}
				}
			}
			// Include comma element to json chars to restore original.
			isJsonChar[44] = true
			// Not found any return error
			if !found {
				return nil, errors.New("Error: Last key not found.")
			}
		}
	}
	// this means not search operation has take place
	// it must be some kinda error or bad format
	if offset == 0 {
		return nil, errors.New("Error: Something went wrong... not sure, maybe bad JSON format...")
	}
	// skip spaces from top.
	for space(json[offset]) {
		offset++
	}
	// If value starts with open braces
	if json[offset] == 91 || json[offset] == 123 {
		// main level indicator.
		level := 0
		// Quote check flag
		inQuote := false
		for i := offset ; i < len(json) ; i ++ {
			// curr is current byte of reading.
			curr := json[i]
			// Just interested with json chars. Other wise continue.
			if !isJsonChar[curr]{
				continue
			}
			if curr == 34 {
				// Just interested with json chars. Other wise continue.
				if json[i - 1] == 92 {
					continue
				}
				inQuote = !inQuote
				continue
			}
			if inQuote {
				continue
			}else{
				if curr == 91 || curr == 123 {
					level++
				}
				if curr == 93 || curr == 125 {
					level--
					if level == 0 {
						// Close brace found in same level with start.
						// Return all of it.
						return json[offset:i + 1], nil
					}
					continue
				}
				continue
			}
			continue
		}
	}else{
		// If value starts with quote
		if json[offset] == 34 {
			inQuote := false
			for i := offset ;  i < len(json) ; i ++ {
				curr := json[i]
				// quote
				// find ending quote
				if curr == 34 {
					// just interested with json chars. Other wise continue.
					if json[i - 1] == 92 {
						continue
					}
					if inQuote {
						// Strip quotes and return.
						return json[offset + 1:i], nil
					}
					inQuote = !inQuote
					continue
				}
			}
		}else{
			// If value starts without quote
			for i := offset ;  i < len(json) ; i ++ {
				if isJsonChar[json[i]] {
					// strip others and return value.
					return json[offset:i], nil
				}
			}
		}
	}
	// This means not search operation has take place
	// not any formatting operation has take place
	// it must be some kinda error or bad format
	return nil, errors.New("Error: Something went wrong... not sure, maybe bad JSON format...")
}

func GetString(json []byte, path ... string) (string, error){
	val, done := Get(json, path...)
	return string(val), done
}

func GetInt(json []byte, path ... string) (int, error){
	val, err := GetString(json, path...)
	if err != nil {
		return -1, err
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return -1, fmt.Errorf("Cast Error: value '%v' can not cast to int.", val)
	}
	return intVal, nil
}

func GetFloat(json []byte, path ... string) (float64, error){
	val, err := GetString(json, path...)
	if err != nil {
		return -1, err
	}
	floatVal, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return -1, fmt.Errorf("Cast Error: value '%v' can not cast to float64.", val)
	}
	return floatVal, nil
}

func GetBool(json []byte, path ... string) (bool, error){
	val, err := GetString(json, path...)
	if err != nil {
		return false, err
	}
	if val == "true" {
		return true, nil
	}
	if val == "false" {
		return false, nil
	}
	return false, fmt.Errorf("Cast Error: value '%v' can not cast to bool.", val)
}

func GetStringArray(json []byte, path ... string) ([]string, error){
	val, err := GetString(json, path...)
	if err != nil {
		return nil, err
	}
	lena := len(val)
	if lena < 2 {
		return nil, fmt.Errorf("Cast Error: value '%v' can not cast to []string.", val)
	}
	if val[0] == '[' && val[lena - 1] == ']' {
		newArray := make([]string, 0, 16)
		start := 1
		inQuote := false
		for i := 1 ; i < lena - 1 ; i ++ {
			curr := val[i]
			if curr == 34 || curr == 44 {
				if curr == 34 {
					// escape character control
					if val[i - 1] == 92 {
						continue
					}
					inQuote = !inQuote
					continue
				}
				if inQuote {
					continue
				}else{
					if curr == 44 {
						newArray = append(newArray, trimSpace(val, start, i))
						start = i + 1
					}
				}
			}
		}
		newArray = append(newArray, trimSpace(val, start, lena - 2))
		return newArray, nil
	}else{
		return nil, fmt.Errorf("Cast Error: value '%v' can not cast to []string.", val)
	}
}

func GetIntArray(json []byte, path ... string) ([]int, error){
	val, err := GetString(json, path...)
	if err != nil {
		return nil, err
	}
	lena := len(val)
	if lena < 2 {
		return nil, fmt.Errorf("Cast Error: value '%v' can not cast to []int.", val)
	}
	if val[0] == '[' && val[lena - 1] == ']' {
		newArray := make([]int, 0, 16)
		start := 1
		inQuote := false
		for i := 1 ; i < lena - 1 ; i ++ {
			curr := val[i]
			if curr == 34 || curr == 44 {
				if curr == 34 {
					// escape character control
					if val[i - 1] == 92 {
						continue
					}
					inQuote = !inQuote
					continue
				}
				if inQuote {
					continue
				}else{
					if curr == 44 {
						num, err := strconv.Atoi(trimSpace(val, start, i))
						if err != nil {
							return nil,  fmt.Errorf("Cast Error: value '%v' can not cast to int.", trimSpace(val, start, i))
						}
						newArray = append(newArray, num)
						start = i + 1
					}
				}
			}
		}

		num, err := strconv.Atoi(trimSpace(val, start, lena - 2))
		if err != nil {
			return nil,  fmt.Errorf("Cast Error: value '%v' can not cast to int.", trimSpace(val, start, lena - 2))
		}
		newArray = append(newArray, num)
		return newArray, nil
	}else{
		return nil, fmt.Errorf("Cast Error: value '%v' can not cast to []int.", val)
	}
}

func GetFloatArray(json []byte, path ... string) ([]float64, error){
	val, err := GetString(json, path...)
	if err != nil {
		return nil, err
	}
	lena := len(val)
	if lena < 2 {
		return nil, fmt.Errorf("Cast Error: value '%v' can not cast to []float64.", val)
	}
	if val[0] == '[' && val[lena - 1] == ']' {
		newArray := make([]float64, 0, 16)
		start := 1
		inQuote := false
		for i := 1 ; i < lena - 1 ; i ++ {
			curr := val[i]
			if curr == 34 || curr == 44 {
				if curr == 34 {
					// escape character control
					if val[i - 1] == 92 {
						continue
					}
					inQuote = !inQuote
					continue
				}
				if inQuote {
					continue
				}else{
					if curr == 44 {
						num, err := strconv.ParseFloat(trimSpace(val, start, i), 64)
						if err != nil {
							return nil,  fmt.Errorf("Cast Error: value '%v' can not cast to float64.", trimSpace(val, start, i))
						}
						newArray = append(newArray, num)
						start = i + 1
					}
				}
			}
		}

		num, err := strconv.ParseFloat(trimSpace(val, start, lena - 2), 64)
		if err != nil {
			return nil,  fmt.Errorf("Cast Error: value '%v' can not cast to float64.", trimSpace(val, start, lena - 2))
		}
		newArray = append(newArray, num)
		return newArray, nil
	}else{
		return nil, fmt.Errorf("Cast Error: value '%v' can not cast to []float64.", val)
	}
}

func GetBoolArray(json []byte, path ... string) ([]bool, error){
	val, err := GetString(json, path...)
	if err != nil {
		return nil, err
	}
	lena := len(val)
	if lena < 2 {
		return nil, fmt.Errorf("Cast Error: value '%v' can not cast to []bool.", val)
	}
	if val[0] == '[' && val[lena - 1] == ']' {
		newArray := make([]bool, 0, 16)
		start := 1
		inQuote := false
		for i := 1 ; i < lena - 1 ; i ++ {
			curr := val[i]
			if curr == 34 || curr == 44 {
				if curr == 34 {
					// escape character control
					if val[i - 1] == 92 {
						continue
					}
					inQuote = !inQuote
					continue
				}
				if inQuote {
					continue
				}else{
					if curr == 44 {
						val := trimSpace(val, start, i)
						if val == "true" || val == "false" {
							if val == "true"{
								newArray = append(newArray, true)
								start = i + 1
							}else{
								newArray = append(newArray, false)
								start = i + 1
							}
						}else{
							return nil,  fmt.Errorf("Cast Error: value '%v' can not cast to bool.", val)
						}
					}
				}
			}
		}
		val := trimSpace(val, start, lena - 2)
		if val == "true" || val == "false" {
			if val == "true"{
				newArray = append(newArray, true)
			}else{
				newArray = append(newArray, false)
			}
		}else{
			return nil,  fmt.Errorf("Cast Error: value '%v' can not cast to bool.", val)
		}
		return newArray, nil
	}else{
		return nil, fmt.Errorf("Cast Error: value '%v' can not cast to []bool.", val)
	}
}


func Set(json []byte, newValue []byte, path ... string) ([]byte, error){
	if len(path) == 0 {
		return nil, errors.New("Error: Path can not be null.")
	}
	if len(newValue) == 0 {
		return nil, errors.New("Error: New Value can not be null.")
	}
	offset := 0
	currentPath := path[0]
	chars := []byte{34, 44, 58, 91, 93, 123, 125}
	isJsonChar := make([]bool, 256)
	for _,v := range chars {
		isJsonChar[v] = true
	}
	for space(json[offset]) {
		offset++
	}
	braceType := json[offset]

	for k := 0 ; k < len(path) ; k ++ {
		if braceType == 91 {
			arrayNumber, err := strconv.Atoi(currentPath)
			if err != nil {
				return json, errors.New("Error: Index Expected.")
			}
			done := false
			if arrayNumber == 0 {
				offset++
				for i := offset; i < len(json) ; i ++ {
					curr := json[i]
					if curr == 123 {
						braceType = curr
						if k != len(path) - 1{
							currentPath = path[k + 1]
						}
						offset = i
						done = true
						break
					}
					if curr == 91 {
						braceType = curr
						if k != len(path) - 1{
							currentPath = path[k + 1]
						}
						offset = i + 1
						done = true
						break
					}
					if !space(curr){
						done = true
						break
					}
				}
			}else{
				level := 0
				inQuote := false
				found := false
				indexCount := 0
				// not interested with column to this level
				isJsonChar[58] = false
				for i := offset ; i < len(json) ; i ++ {
					curr := json[i]
					if !isJsonChar[curr]{
						continue
					}
					if curr == 34 {
						if json[i - 1] == 92 {
							continue
						}
						inQuote = !inQuote
						continue
					}
					if inQuote {
						continue
					}else{
						if curr == 91 || curr == 123{
							if found {
								level++
								braceType = curr
								currentPath = path[k + 1]
								found = false
								done = true
								break
							}
							level++
							continue
						}
						if curr == 93 || curr == 125 {
							level--
							if level < 1 {
								done = false
								break
							}
							continue
						}
						if !found {
							if level == 1 {
								if curr == 44 {
									indexCount++
									if indexCount == arrayNumber {
										offset = i + 1
										if k == len(path) - 1{
											done = true
											break
										}
										found = true
										continue
									}
									continue
								}
								continue
							}
							continue
						}
						continue
					}
				}
				// interested with column to this level
				isJsonChar[58] = true
			}
			if !done {
				return json, errors.New("Error: Index out of range")
			}
		}else{
			inQuote := false
			found := false
			start := 0
			end := 0
			level := k
			// not interested with comma to this level
			isJsonChar[44] = false
			for i := offset ; i < len(json) ; i ++ {
				curr := json[i]
				if !isJsonChar[curr]{
					continue
				}
				if curr == 34 {
					inQuote = !inQuote
					if found {
						continue
					}
					if level != k + 1 {
						continue
					}
					if inQuote {
						start = i + 1
						continue
					}
					end = i
					continue
				}
				if inQuote {
					continue
				}else{
					if curr == 91 {
						if found {
							braceType = curr
							currentPath = path[k + 1]
							break
						}
						level++
						continue
					}
					if curr == 123 {
						if found {
							k++
							level++
							currentPath = path[k]
							found = false
							continue
						}
						level++
						continue
					}
					// close brace
					if curr == 93 || curr == 125 {
						level--
						continue
					}
					// column
					if level == k + 1 {
						if curr == 58 {
							if compare(json, start, end, currentPath) {
								offset = i + 1
								found = true
								if k == len(path) - 1{
									isJsonChar[44] = true
									break
								}else{
									continue
								}
							}
							// interested with comma to this level
							isJsonChar[44] = true
							// not interested with column to this level
							isJsonChar[58] = false
							// little jump algorithm :{} -> ,
							for j := i ;  j < len(json) ; j ++ {
								curr := json[j]
								if !isJsonChar[curr]{
									continue
								}
								// quote
								if curr == 34 {
									if json[j - 1] == 92 {
										continue
									}
									inQuote = !inQuote
									continue
								}
								if inQuote {
									continue
								}else{
									if curr == 91 || curr == 123 {
										level++
										continue
									}
									if curr == 93 || curr == 125 {
										level--
										continue
									}
									// comma
									if curr == 44 {
										if level == k + 1 {
											i = j
											break
										}
										continue
									}
									continue
								}

							}
							// not interested with comma to this level
							isJsonChar[44] = false
							// interested with column to this level
							isJsonChar[58] = true
							continue
						}
						continue
					}
				}
			}
			isJsonChar[44] = true
			if !found {
				return json, errors.New("Error: Last key not found.")
			}
		}
	}
	if offset == 0 {
		return json, errors.New("Error: Non")
	}
	for space(json[offset]) {
		offset++
	}
	// starts with { [
	if json[offset] == 91 || json[offset] == 123 {
		level := 0
		inQuote := false
		for i := offset ; i < len(json) ; i ++ {
			curr := json[i]
			if !isJsonChar[curr]{
				continue
			}
			if curr == 34 {
				// escape character control
				if json[i - 1] == 92 {
					continue
				}
				inQuote = !inQuote
				continue
			}
			if inQuote {
				continue
			}else{
				if curr == 91 || curr == 123 {
					level++
				}
				if curr == 93 || curr == 125 {
					level--
					if level == 0 {
						return replace(json, newValue, offset, i + 1), nil
					}
					continue
				}
				continue
			}
			continue
		}
	}else{
		// starts with quote
		if json[offset] == 34 {
			inQuote := false
			for i := offset ;  i < len(json) ; i ++ {
				curr := json[i]
				// quote
				if curr == 34 {
					// escape character control
					if json[i - 1] == 92 {
						continue
					}
					if inQuote {
						return replace(json, newValue, offset, i + 1), nil
					}
					inQuote = !inQuote
					continue
				}
			}
		}else{
			// starts without quote
			for i := offset ;  i < len(json) ; i ++ {
				if isJsonChar[json[i]] {
					return replace(json, newValue, offset, i), nil
				}
			}
		}
	}
	return nil, errors.New("Error: Non 2")
}

func SetString(json []byte, newValue string, path ... string) ([]byte, error){
	if newValue[0] != 34 && newValue[len(newValue) - 1] != 34 {
		return Set(json, []byte(`"` + newValue + `"`), path...)
	}
	return Set(json, []byte(newValue), path...)
}

func SetInt(json []byte, newValue int, path ... string) ([]byte, error){
	return Set(json, []byte(strconv.Itoa(newValue)), path...)
}

func SetFloat(json []byte, newValue float64, path ... string) ([]byte, error){
	return Set(json, []byte(strconv.FormatFloat(newValue, 'e', -1, 64)), path...)
}

func SetBool(json []byte, newValue bool, path ... string) ([]byte, error){
	if newValue {
		return Set(json, []byte("true"), path...)
	}
	return Set(json, []byte("false"), path...)
}


func SetKey(json []byte, newValue []byte, path ... string) ([]byte, error){
	if len(path) == 0 {
		return json, errors.New("Error: Path can not be null.")
	}
	if len(newValue) == 0 {
		return json, errors.New("Error: New Value can not be null.")
	}
	for _, v := range newValue {
		if v  == 34 {
			return json, errors.New("Error: Key can not contain quote symbol.")
		}
	}
	offset := 0
	currentPath := path[0]
	chars := []byte{34, 44, 58, 91, 93, 123, 125}
	isJsonChar := make([]bool, 256)
	for _,v := range chars {
		isJsonChar[v] = true
	}
	for space(json[offset]) {
		offset++
	}
	braceType := json[offset]

	for k := 0 ; k < len(path) ; k ++ {
		if braceType == 91 {
			arrayNumber, err := strconv.Atoi(currentPath)
			if err != nil {
				return json, errors.New("Error: Index Expected.")
			}
			done := false
			if arrayNumber == 0 {
				offset++
				for i := offset; i < len(json) ; i ++ {
					curr := json[i]
					if curr == 123 {
						braceType = curr
						if k != len(path) - 1{
							currentPath = path[k + 1]
						}
						offset = i
						done = true
						break
					}
					if curr == 91 {
						braceType = curr
						if k != len(path) - 1{
							currentPath = path[k + 1]
						}
						offset = i + 1
						done = true
						break
					}
					if !space(curr){
						done = true
						break
					}
				}
			}else{
				level := 0
				inQuote := false
				found := false
				indexCount := 0
				// not interested with column to this level
				isJsonChar[58] = false
				for i := offset ; i < len(json) ; i ++ {
					curr := json[i]
					if !isJsonChar[curr]{
						continue
					}
					if curr == 34 {
						if json[i - 1] == 92 {
							continue
						}
						inQuote = !inQuote
						continue
					}
					if inQuote {
						continue
					}else{
						if curr == 91 || curr == 123{
							if found {
								level++
								braceType = curr
								currentPath = path[k + 1]
								found = false
								done = true
								break
							}
							level++
							continue
						}
						if curr == 93 || curr == 125 {
							level--
							if level < 1 {
								done = false
								break
							}
							continue
						}
						if !found {
							if level == 1 {
								if curr == 44 {
									indexCount++
									if indexCount == arrayNumber {
										offset = i + 1
										if k == len(path) - 1{
											done = true
											return json, errors.New("Error: Last value must be a key value,  not an array index.")
										}
										found = true
										continue
									}
									continue
								}
								continue
							}
							continue
						}
						continue
					}
				}
				// interested with column to this level
				isJsonChar[58] = true
			}
			if !done {
				return json, errors.New("Error: Index out of range")
			}
		}else{
			inQuote := false
			found := false
			start := 0
			end := 0
			level := k
			// not interested with comma to this level
			isJsonChar[44] = false
			for i := offset ; i < len(json) ; i ++ {
				curr := json[i]
				if !isJsonChar[curr]{
					continue
				}
				if curr == 34 {
					inQuote = !inQuote
					if found {
						continue
					}
					if level != k + 1 {
						continue
					}
					if inQuote {
						start = i + 1
						continue
					}
					end = i
					continue
				}
				if inQuote {
					continue
				}else{
					if curr == 91 {
						if found {
							braceType = curr
							currentPath = path[k + 1]
							break
						}
						level++
						continue
					}
					if curr == 123 {
						if found {
							k++
							level++
							currentPath = path[k]
							found = false
							continue
						}
						level++
						continue
					}
					// close brace
					if curr == 93 || curr == 125 {
						level--
						continue
					}
					// column
					if level == k + 1 {
						if curr == 58 {
							if compare(json, start, end, currentPath) {
								offset = i + 1
								found = true
								if k == len(path) - 1{
									isJsonChar[44] = true
									return replace(json, newValue, start, end), nil
									break
								}else{
									continue
								}
							}
							// interested with comma to this level
							isJsonChar[44] = true
							// not interested with column to this level
							isJsonChar[58] = false
							// little jump algorithm :{} -> ,
							for j := i ;  j < len(json) ; j ++ {
								curr := json[j]
								if !isJsonChar[curr]{
									continue
								}
								// quote
								if curr == 34 {
									if json[j - 1] == 92 {
										continue
									}
									inQuote = !inQuote
									continue
								}
								if inQuote {
									continue
								}else{
									if curr == 91 || curr == 123 {
										level++
										continue
									}
									if curr == 93 || curr == 125 {
										level--
										continue
									}
									// comma
									if curr == 44 {
										if level == k + 1 {
											i = j
											break
										}
										continue
									}
									continue
								}

							}
							// not interested with comma to this level
							isJsonChar[44] = false
							// interested with column to this level
							isJsonChar[58] = true
							continue
						}
						continue
					}
				}
			}
			isJsonChar[44] = true
			if !found {
				return json, errors.New("Error: Last key not found.")
			}
		}
	}
	return json, errors.New("Error: Something went wrong... not sure.")
}

func SetStringKey(json []byte, newValue string, path ... string) ([]byte, error){
	return SetKey(json, []byte(newValue), path...)
}

func replace(json, newValue []byte, start, end int) []byte {
	newJson := make([]byte, 0, len(json) - end + start + len(newValue))
	newJson = append(newJson, json[:start]...)
	newJson = append(newJson, newValue...)
	newJson = append(newJson, json[end:]...)
	return newJson
}

func trimSpace(str string, start, eoe int) string {
	for space(str[start]){
		start++
	}
	end := start
	for !space(str[end]) && end < eoe {
		end++
	}
	return str[start:end]
}

func compare(json []byte, start, end int , key string) bool{
	if len(key) != end - start {
		return false
	}
	for i := 0 ; i < len(key) ; i ++ {
		if key[i] != json[start + i] {
			return false
		}
	}
	return true
}

func space(curr byte) bool{
	// space
	if curr == 32 {
		return true
	}
	// tab
	if curr == 9 {
		return true
	}
	// new line NL
	if curr == 10 {
		return true
	}
	// return CR
	if curr == 13 {
		return true
	}
	return false
}