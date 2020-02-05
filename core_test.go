package jin

import (
	test "jin/test"
	"strconv"
	"strings"
	"testing"
)

var (
	json   []byte
	paths  [][]string
	values []string
	tests  []string
)

func init() {
	tests = test.Dir(test.GetCurrentDir() + test.Sep() + "tests")
}

func InitValues(t *testing.T, flat bool, scenario string, fileName string) {
	test.WriteFile("test/test-json.json", test.ReadFile(fileName))
	json = test.ReadFile("test/test-json.json")
	if flat {
		json = Flatten(json)
	}
	str, err := test.ExecuteNode(scenario)
	if err != nil {
		t.Errorf("Init Error E:%v, S:%v\n", err, str)
		return
	}
	pathFile := string(test.ReadFile("test/test-json-paths.json"))
	valueFile := string(test.ReadFile("test/test-json-values.json"))
	if pathFile == "" || valueFile == "" {
		paths = make([][]string, 0)
		values = make([]string, 0)
		t.Logf("SKIPED.\n")
		return
	}
	newPaths := strings.Split(pathFile, "\n")
	newValues := strings.Split(valueFile, "\n")
	paths = make([][]string, 0, len(newPaths))
	values = make([]string, 0, len(newValues))
	if len(newPaths) == 0 {
		t.Logf("Paths length is zero.\n")
		return
	}
	if len(newValues) == 0 {
		t.Logf("Values length is zero.\n")
		return
	}
	for _, val := range newValues {
		values = append(values, val)
	}
	for _, val := range newPaths {
		paths = append(paths, ParseArray(val))
	}
}

func TestInterperterGet(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, false, "get", file)
		for i, _ := range paths {
			_, start, end, err := core(json, false, paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Test Get), path:%v err:%v\n", paths[i], err)
				return
			}
			value := json[start:end]
			if json[start-1] != 34 {
				value = Flatten(value)
			}
			if string(value) != stripQuotes(values[i]) {
				t.Errorf("Fail (Test Get), not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<  i:%v\n", paths[i], string(value), stripQuotes(values[i]), i)
				return
			}
		}
	}
}

func TestInterperterSet(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, false, "set", file)
		for i, _ := range paths {
			value, err := Set(json, []byte(`"test-string"`), paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Set), path:%v err:%v\n", paths[i], err)
				return
			}
			value, err = Get(value, paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Get), path:%v err:%v\n", paths[i], err)
				return
			}
			if string(value) != stripQuotes(values[i]) {
				t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", paths[i], string(value), values[i])
				return
			}
		}
	}
}

func TestInterperterSetKey(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, false, "get", file)
		for i, _ := range paths {
			keyStart, _, _, err1 := core(json, true, paths[i]...)
			if err1 != nil {
				t.Errorf("Total Fail(Core), path:%v\n", paths[i])
				return
			}
			newJson, err2 := SetKey(json, "test-key", paths[i]...)
			// it is a number
			if keyStart == -1 {
				if err2 == nil {
					t.Errorf("It is an element of an array cannot be set a new key %v", paths[i])
					return
				}
			} else {
				if err2 != nil {
					t.Errorf("It is a key it can be set a new key %v", paths[i])
					return
				}
				newPath := make([]string, len(paths[i]))
				copy(newPath, paths[i][:len(paths[i])-1])
				newPath[len(newPath)-1] = "test-key"
				_, start, end, err := core(newJson, false, newPath...)
				if err != nil {
					t.Errorf("Total Fail(Get), path:%v err:%v\n", paths[i], err)
					return
				}
				value := newJson[start:end]
				if newJson[start-1] != 34 {
					value = Flatten(value)
				}
				if string(value) != stripQuotes(values[i]) {
					t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", newPath, string(value), values[i])
					return
				}
			}
		}
	}
}

func TestInterperterAddKV(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, false, "addkv", file)
		for i, _ := range paths {
			value, err := AddKeyValue(json, "test-key", []byte(`"test-value"`), paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Set), path:%v err:%v\n", paths[i], err)
				return
			}
			var start int
			var end int
			var val []byte
			if len(paths[i]) == 0 {
				start = 0
				end = len(value)
				val = value
				val = Flatten(val)
			}else{
				_, start, end, err = core(value, false, paths[i]...)
				if err != nil {
					t.Errorf("Total Fail(Get), path:%v err:%v\n", paths[i], err)
					return
				}
				val = value[start:end]
				if value[start-1] != 34 {
					val = Flatten(val)
				}
			}
			if string(val) != stripQuotes(values[i]) {
				t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", paths[i], string(val), values[i])
				return
			}
		}
	}
}

func TestInterperterAdd(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, false, "add", file)
		for i, _ := range paths {
			value, err := Add(json, []byte(`"test-value"`), paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Set), path:%v err:%v\n", paths[i], err)
				return
			}
			value, err = Get(value, paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Get), path:%v err:%v\n", paths[i], err)
				return
			}
			if string(Flatten(value)) != stripQuotes(values[i]) {
				t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", paths[i], string(Flatten(value)), values[i])
				return
			}
		}
	}
}

func TestInterperterInsert(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, false, "insert", file)
		var err error
		var value []byte
		for i, _ := range paths {
			json, err = Insert(json, 0, []byte(`"test-value"`), paths[i]...)
			if err != nil {
				if err.Error() != EMPTY_ARRAY_ERROR().Error() {
					t.Errorf("Total Fail(Set), path:%v err:%v\n", paths[i], err)
					return
				} else {
					continue
				}
			}
			value, err = Get(json, paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Get), path:%v err:%v\n", paths[i], err)
				return
			}
			if string(Flatten(value)) != stripQuotes(values[i]) {
				t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", paths[i], string(value), values[i])
				return
			}
		}
	}
}

func TestInterperterDeleteKV(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, false, "deleteKV", file)
		for i, _ := range paths {
			value, err := AddKeyValue(json, "test-key", []byte(`"test-value"`), paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Set), path:%v err:%v\n", paths[i], err)
				return
			}
			newPath := make([]string, len(paths[i]))
			copy(newPath, paths[i])
			newPath = append(newPath, "test-key")
			value, err = Delete(value, newPath...)
			if err != nil {
				t.Errorf("Total Fail(Set), path:%v err:%v\n", paths[i], err)
				return
			}
			value, err = Get(value, paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Get), path:%v err:%v\n", paths[i], err)
				return
			}
			if string(Flatten(value)) != stripQuotes(values[i]) {
				t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", paths[i], string(value), values[i])
				return
			}
		}
	}
}

func TestInterperterDeleteV(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, false, "deleteV", file)
		for i, _ := range paths {
			empty, err := IsEmpty(json, paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Empty), path:%v err:%v\n", paths[i], err)
				return
			}
			if empty {
				t.Logf("SKIP(EMPTY ARRAY)")
				return
			}
			value, err := Get(json, paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Get), path:%v err:%v\n", paths[i], err)
				return
			}
			arr := ParseArray(string(value))
			value, err = Add(json, []byte(`"test-value"`), paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Set), path:%v err:%v\n", paths[i], err)
				return
			}
			newPath := make([]string, len(paths[i]))
			copy(newPath, paths[i])
			newPath = append(newPath, strconv.Itoa(len(arr)))
			value, err = Delete(value, newPath...)
			if err != nil {
				t.Errorf("Total Fail(Set), path:%v err:%v\n", paths[i], err)
				return
			}
			value, err = Get(value, paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Get), path:%v err:%v\n", paths[i], err)
				return
			}
			if string(Flatten(value)) != stripQuotes(values[i]) {
				t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", paths[i], string(value), values[i])
				return
			}
		}
	}
}

func TestInterperterArrayIter(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, false, "arrayiter", file)
		for _, path := range paths {
			count := 0
			err := IterateArray(json, func(value []byte) bool {
				newPath := make([]string, len(path))
				copy(newPath, path)
				newPath = append(newPath, strconv.Itoa(count))
				value2, err := Get(json, newPath...)
				if err != nil {
					t.Errorf("Total Fail (Iter Array Get), not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", newPath, string(value2), string(value))
				}
				if string(value) != string(value2) {
					t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", newPath, string(value2), string(value))
					return false
				} else {
					count++
					return true
				}
			}, path...)
			if err != nil {
				t.Errorf("Total Fail(ArrayIter), path:%v err:%v\n", path, err)
				return
			}
		}
	}
}

func TestInterperterKeyValueIter(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, false, "objectiter", file)
		for _, path := range paths {
			err := IterateKeyValue(json, func(key []byte, value []byte) bool {
				newPath := make([]string, len(path))
				copy(newPath, path)
				newPath = append(newPath, string(key))
				value2, err := Get(json, newPath...)
				if err != nil {
					t.Errorf("Total Fail (Iter Key Value Get), not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", newPath, string(value), string(value2))
				}
				if string(value) != string(value2) {
					t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", newPath, string(value), string(value2))
					return false
				} else {
					return true
				}
			}, path...)
			if err != nil {
				t.Errorf("Total Fail(ArrayIter), path:%v err:%v\n", path, err)
				return
			}
		}
	}
}

func TestInterperterGetFlatten(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, true, "get", file)
		for i, _ := range paths {
			_, start, end, err := core(json, false, paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Test Get), path:%v err:%v\n", paths[i], err)
				return
			}
			// t.Logf("val:>%v<\n", string(value))
			value := json[start:end]
			if json[start-1] != 34 {
				value = Flatten(value)
			}
			if string(value) != stripQuotes(values[i]) {
				t.Errorf("Fail (Test Get), not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<  i:%v\n", paths[i], string(value), stripQuotes(values[i]), i)
				return
			}
		}
	}
}

func TestInterperterSetFlatten(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, true, "set", file)
		for i, _ := range paths {
			value, err := Set(json, []byte(`"test-string"`), paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Set), path:%v err:%v\n", paths[i], err)
				return
			}
			value, err = Get(value, paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Get), path:%v err:%v\n", paths[i], err)
				return
			}
			if string(value) != stripQuotes(values[i]) {
				t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", paths[i], string(value), values[i])
				return
			}
		}
	}
}

func TestInterperterSetKeyFlatten(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, true, "get", file)
		for i, _ := range paths {
			keyStart, _, _, err1 := core(json, true, paths[i]...)
			if err1 != nil {
				t.Errorf("Total Fail(Core), path:%v\n", paths[i])
				return
			}
			newJson, err2 := SetKey(json, "test-key", paths[i]...)
			// it is a number
			if keyStart == -1 {
				if err2 == nil {
					t.Errorf("It is an element of an array cannot be set a new key %v", paths[i])
					return
				}
			} else {
				if err2 != nil {
					t.Errorf("It is a key it can be set a new key %v", paths[i])
					return
				}
				newPath := make([]string, len(paths[i]))
				copy(newPath, paths[i][:len(paths[i])-1])
				newPath[len(newPath)-1] = "test-key"
				_, start, end, err := core(newJson, false, newPath...)
				if err != nil {
					t.Errorf("Total Fail(Get), path:%v err:%v\n", paths[i], err)
					return
				}
				value := newJson[start:end]
				if newJson[start-1] != 34 {
					value = Flatten(value)
				}
				if string(value) != stripQuotes(values[i]) {
					t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", newPath, string(value), values[i])
					return
				}
			}
		}
	}
}

func TestInterperterAddKVFlatten(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, true, "addkv", file)
		for i, _ := range paths {
			value, err := AddKeyValue(json, "test-key", []byte(`"test-value"`), paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Set), path:%v err:%v\n", paths[i], err)
				return
			}
			value, err = Get(value, paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Get), path:%v err:%v\n", paths[i], err)
				return
			}
			if string(Flatten(value)) != stripQuotes(values[i]) {
				t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", paths[i], string(value), values[i])
				return
			}
		}
	}
}

func TestInterperterAddFlatten(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, true, "add", file)
		for i, _ := range paths {
			value, err := Add(json, []byte(`"test-value"`), paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Set), path:%v err:%v\n", paths[i], err)
				return
			}
			value, err = Get(value, paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Get), path:%v err:%v\n", paths[i], err)
				return
			}
			if string(Flatten(value)) != stripQuotes(values[i]) {
				t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", paths[i], string(Flatten(value)), values[i])
				return
			}
		}
	}
}

func TestInterperterInsertFlatten(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, true, "insert", file)
		var err error
		var value []byte
		for i, _ := range paths {
			json, err = Insert(json, 0, []byte(`"test-value"`), paths[i]...)
			if err != nil {
				if err.Error() != EMPTY_ARRAY_ERROR().Error() {
					t.Errorf("Total Fail(Insert), path:%v err:%v\n", paths[i], err)
					return
				} else {
					continue
				}
			}
			value, err = Get(json, paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Insert Get), path:%v err:%v\n", paths[i], err)
				return
			}
			if string(Flatten(value)) != stripQuotes(values[i]) {
				t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", paths[i], string(value), values[i])
				return
			}
		}
	}
}

func TestInterperterDeleteKVFlatten(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, true, "deleteKV", file)
		for i, _ := range paths {
			value, err := AddKeyValue(json, "test-key", []byte(`"test-value"`), paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Set), path:%v err:%v\n", paths[i], err)
				return
			}
			newPath := make([]string, len(paths[i]))
			copy(newPath, paths[i])
			newPath = append(newPath, "test-key")
			value, err = Delete(value, newPath...)
			if err != nil {
				t.Errorf("Total Fail(Set), path:%v err:%v\n", paths[i], err)
				return
			}
			value, err = Get(value, paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Get), path:%v err:%v\n", paths[i], err)
				return
			}
			if string(Flatten(value)) != stripQuotes(values[i]) {
				t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", paths[i], string(value), values[i])
				return
			}
		}
	}
}

func TestInterperterDeleteVFlatten(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, true, "deleteV", file)
		for i, _ := range paths {
			value, err := Get(json, paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Get), path:%v err:%v\n", paths[i], err)
				return
			}
			arr := ParseArray(string(value))
			value, err = Add(json, []byte(`"test-value"`), paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Set), path:%v err:%v\n", paths[i], err)
				return
			}
			newPath := make([]string, len(paths[i]))
			copy(newPath, paths[i])
			newPath = append(newPath, strconv.Itoa(len(arr)))
			value, err = Delete(value, newPath...)
			if err != nil {
				t.Errorf("Total Fail(Set), path:%v err:%v\n", paths[i], err)
				return
			}
			value, err = Get(value, paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Get), path:%v err:%v\n", paths[i], err)
				return
			}
			if string(Flatten(value)) != stripQuotes(values[i]) {
				t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", paths[i], string(value), values[i])
				return
			}
		}
	}
}

func TestInterperterArrayIterFlatten(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, true, "arrayiter", file)
		for _, path := range paths {
			count := 0
			err := IterateArray(json, func(value []byte) bool {
				newPath := make([]string, len(path))
				copy(newPath, path)
				newPath = append(newPath, strconv.Itoa(count))
				value2, err := Get(json, newPath...)
				if err != nil {
					t.Errorf("Total Fail (Iter Array Get), not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", newPath, string(value2), string(value))
				}
				if string(value) != string(value2) {
					t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", newPath, string(value2), string(value))
					return false
				} else {
					count++
					return true
				}
			}, path...)
			if err != nil {
				t.Errorf("Total Fail(ArrayIter), path:%v err:%v\n", path, err)
				return
			}
		}
	}
}

func TestInterperterKeyValueIterFlatten(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, true, "objectiter", file)
		for _, path := range paths {
			err := IterateKeyValue(json, func(key []byte, value []byte) bool {
				newPath := make([]string, len(path))
				copy(newPath, path)
				newPath = append(newPath, string(key))
				value2, err := Get(json, newPath...)
				if err != nil {
					t.Errorf("Total Fail (Iter Key Value Get), not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", newPath, string(value), string(value2))
				}
				if string(value) != string(value2) {
					t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", newPath, string(value), string(value2))
					return false
				} else {
					return true
				}
			}, path...)
			if err != nil {
				t.Errorf("Total Fail(ArrayIter), path:%v err:%v\n", path, err)
				return
			}
		}
	}
}

func TestParserGet(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, false, "get", file)
		prs, err := Parse(json)
		if err != nil {
			t.Errorf("Total Fail(Parse Get), err:%v\n", err)
			return
		}
		for i, _ := range paths {
			value, err := prs.Get(paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Test Parse Get), path:%v err:%v\n", paths[i], err)
				return
			}
			if len(value) > 1 {
				if value[0] == 91 && value[len(value) - 1] == 93 || value[0] == 123 && value[len(value) - 1] == 125 {
					value = Flatten(value)
				}
			}
			if string(value) != stripQuotes(values[i]) {
				t.Errorf("Fail (Test Get), not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<  i:%v\n", paths[i], string(value), stripQuotes(values[i]), i)
				return
			}
		}
	}
}

func TestParserSet(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, false, "set", file)
		prs, err := Parse(json)
		if err != nil {
			t.Errorf("Total Fail(Parse Set), err:%v\n", err)
			return
		}
		for i, _ := range paths {
			err := prs.Set([]byte(`"test-string"`), paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Test Parse Set), path:%v err:%v\n", paths[i], err)
				return
			}
			value, err := prs.Get(paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Test Parse Get), path:%v err:%v\n", paths[i], err)
				return
			}
			if len(value) > 1 {
				if value[0] == 91 && value[len(value) - 1] == 93 || value[0] == 123 && value[len(value) - 1] == 125 {
					value = Flatten(value)
				}
			}
			if string(value) != stripQuotes(values[i]) {
				t.Errorf("Fail (Test Parse Set), not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<  i:%v\n", paths[i], string(value), stripQuotes(values[i]), i)
				return
			}
		}
	}
}

func TestParserSetKey(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, false, "get", file)
		for i, _ := range paths {
			prs, err := Parse(json)
			if err != nil {
				t.Errorf("Total Fail(Parse SetKey), err:%v\n", err)
				return
			}
			value, err := prs.Get(paths[i][:len(paths[i]) - 1]...)
			if err != nil {
				t.Errorf("Total Fail(Test Parse SetKey Get), path:%v err:%v\n", paths[i], err)
				return
			}
			if len(value) > 1 && len(paths[i]) > 1{
				if value[0] == 123 && value[len(value) - 1] == 125 {
					err := prs.SetKey("test-key", paths[i]...)
					if err != nil {
						t.Errorf("Total Fail(Test Parse SetKey), path:%v err:%v\n", paths[i], err)
						return
					}
					value2, err := prs.Get(append(paths[i][:len(paths[i]) - 1], "test-key")...)
					if err != nil {
						t.Errorf("Total Fail(Test Parse SetKey Get), path:%v err:%v\n", paths[i], err)
						return
					}
					value, err := prs.Get(paths[i]...)
					if err != nil {
						t.Errorf("Total Fail(Test Parse SetKey Get2), path:%v err:%v\n", paths[i], err)
						return
					}
					if string(value) != string(value2) {
						t.Errorf("Fail (Test Parse SetKey), not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<  i:%v\n", paths[i], string(value), string(value2), i)
						return
					}
				}
			}
		}
	}
}

func TestParserAddKV(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, false, "addkv", file)
		for i, _ := range paths {
			prs, err := Parse(json)
			if err != nil {
				t.Errorf("Total Fail(Parse AddKV), err:%v\n", err)
				return
			}
			err = prs.AddKeyValue("test-key", []byte(`"test-value"`), paths[i]...)
			value, err := prs.Get(paths[i]...)
			if string(Flatten(value)) != string(values[i]) {
				t.Errorf("Fail (Test Parse AddKV), not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<  i:%v\n", paths[i], string(Flatten(value)), string(values[i]), i)
			}
		}
	}
}

func TestParserAdd(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, false, "add", file)
		for i, _ := range paths {
			prs, err := Parse(json)
			if err != nil {
				t.Errorf("Total Fail(Parse Add), err:%v\n", err)
				return
			}
			err = prs.Add([]byte(`"test-value"`), paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Add), path:%v err:%v\n", paths[i], err)
				return
			}
			value, err := prs.Get(paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Add), path:%v err:%v\n", paths[i], err)
				return
			}
			if string(Flatten(value)) != values[i] {
				t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", paths[i], string(Flatten(value)), values[i])
				return
			}
		}
	}
}

func TestParserInsert(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, false, "insert", file)
		var value []byte
		prs, err := Parse(json)
		if err != nil {
			t.Errorf("Total Fail(Parse Insert), err:%v\n", err)
			return
		}
		for i, _ := range paths {
			empty, err := IsEmpty(json, paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Empty), path:%v err:%v\n", paths[i], err)
				return
			}
			if empty {
				t.Logf("SKIP(EMPTY ARRAY)")
				return
			}
			err = prs.Insert(0, []byte(`"test-value"`), paths[i]...)
			if err != nil {
				if err.Error() != EMPTY_ARRAY_ERROR().Error() {
					t.Errorf("Total Fail(Insert), path:%v err:%v\n", paths[i], err)
					return
				} else {
					continue
				}
			}
			value, err = prs.Get(paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Get Insert), path:%v err:%v\n", paths[i], err)
				return
			}
			if string(Flatten(value)) != stripQuotes(values[i]) {
				t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", paths[i], string(value), values[i])
				return
			}
		}
	}
}

func TestParserDeleteKV(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, false, "deleteKV", file)
		for i, _ := range paths {
			prs, err := Parse(json)
			if err != nil {
				t.Errorf("Total Fail(Parse Insert), err:%v\n", err)
				return
			}
			err = prs.AddKeyValue("test-key", []byte(`"test-value"`), paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Set), path:%v err:%v\n", paths[i], err)
				return
			}
			newPath := make([]string, len(paths[i]))
			copy(newPath, paths[i])
			newPath = append(newPath, "test-key")
			err = prs.Delete(newPath...)
			if err != nil {
				t.Errorf("Total Fail(Set), path:%v err:%v\n", paths[i], err)
				return
			}
			value, err := prs.Get(paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Get), path:%v err:%v\n", paths[i], err)
				return
			}
			if string(Flatten(value)) != stripQuotes(values[i]) {
				t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", paths[i], string(value), values[i])
				return
			}
		}
	}
}

func TestParserDeleteV(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, false, "deleteV", file)
		for i, _ := range paths {
			empty, err := IsEmpty(json, paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Empty), path:%v err:%v\n", paths[i], err)
				return
			}
			if empty {
				t.Logf("SKIP(EMPTY ARRAY)")
				return
			}
			prs, err := Parse(json)
			if err != nil {
				t.Errorf("Total Fail(Parse Insert), err:%v\n", err)
				return
			}
			value, err := prs.Get(paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Get), path:%v err:%v\n", paths[i], err)
				return
			}
			arr := ParseArray(string(value))
			err = prs.Add([]byte(`"test-value"`), paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Set), path:%v err:%v\n", paths[i], err)
				return
			}
			newPath := make([]string, len(paths[i]))
			copy(newPath, paths[i])
			newPath = append(newPath, strconv.Itoa(len(arr)))
			err = prs.Delete(newPath...)
			if err != nil {
				t.Errorf("Total Fail(Set), path:%v err:%v\n", paths[i], err)
				return
			}
			value, err = prs.Get(paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Get), path:%v err:%v\n", paths[i], err)
				return
			}
			if string(Flatten(value)) != stripQuotes(values[i]) {
				t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", paths[i], string(value), values[i])
				return
			}
		}
	}
}

func TestParserGetFlatten(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, true, "get", file)
		prs, err := Parse(json)
		if err != nil {
			t.Errorf("Total Fail(Parse Get), err:%v\n", err)
			return
		}
		for i, _ := range paths {
			value, err := prs.Get(paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Test Parse Get), path:%v err:%v\n", paths[i], err)
				return
			}
			if len(value) > 1 {
				if value[0] == 91 && value[len(value) - 1] == 93 || value[0] == 123 && value[len(value) - 1] == 125 {
					value = Flatten(value)
				}
			}
			if string(value) != stripQuotes(values[i]) {
				t.Errorf("Fail (Test Get), not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<  i:%v\n", paths[i], string(value), stripQuotes(values[i]), i)
				return
			}
		}
	}
}

func TestParserSetFlatten(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, true, "set", file)
		prs, err := Parse(json)
		if err != nil {
			t.Errorf("Total Fail(Parse Set), err:%v\n", err)
			return
		}
		for i, _ := range paths {
			err := prs.Set([]byte(`"test-string"`), paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Test Parse Set), path:%v err:%v\n", paths[i], err)
				return
			}
			value, err := prs.Get(paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Test Parse Get), path:%v err:%v\n", paths[i], err)
				return
			}
			if len(value) > 1 {
				if value[0] == 91 && value[len(value) - 1] == 93 || value[0] == 123 && value[len(value) - 1] == 125 {
					value = Flatten(value)
				}
			}
			if string(value) != stripQuotes(values[i]) {
				t.Errorf("Fail (Test Parse Set), not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<  i:%v\n", paths[i], string(value), stripQuotes(values[i]), i)
				return
			}
		}
	}
}

func TestParserSetKeyFlatten(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, true, "get", file)
		for i, _ := range paths {
			prs, err := Parse(json)
			if err != nil {
				t.Errorf("Total Fail(Parse SetKey), err:%v\n", err)
				return
			}
			value, err := prs.Get(paths[i][:len(paths[i]) - 1]...)
			if err != nil {
				t.Errorf("Total Fail(Test Parse SetKey Get), path:%v err:%v\n", paths[i], err)
				return
			}
			if len(value) > 1 && len(paths[i]) > 1{
				if value[0] == 123 && value[len(value) - 1] == 125 {
					err := prs.SetKey("test-key", paths[i]...)
					if err != nil {
						t.Errorf("Total Fail(Test Parse SetKey), path:%v err:%v\n", paths[i], err)
						return
					}
					value2, err := prs.Get(append(paths[i][:len(paths[i]) - 1], "test-key")...)
					if err != nil {
						t.Errorf("Total Fail(Test Parse SetKey Get), path:%v err:%v\n", paths[i], err)
						return
					}
					value, err := prs.Get(paths[i]...)
					if err != nil {
						t.Errorf("Total Fail(Test Parse SetKey Get2), path:%v err:%v\n", paths[i], err)
						return
					}
					if string(value) != string(value2) {
						t.Errorf("Fail (Test Parse SetKey), not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<  i:%v\n", paths[i], string(value), string(value2), i)
						return
					}
				}
			}
		}
	}
}


func TestParserAddKVFlatten(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, true, "addkv", file)
		for i, _ := range paths {
			prs, err := Parse(json)
			if err != nil {
				t.Errorf("Total Fail(Parse AddKV), err:%v\n", err)
				return
			}
			err = prs.AddKeyValue("test-key", []byte(`"test-value"`), paths[i]...)
			value, err := prs.Get(paths[i]...)
			if string(Flatten(value)) != string(values[i]) {
				t.Errorf("Fail (Test Parse AddKV), not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<  i:%v\n", paths[i], string(Flatten(value)), string(values[i]), i)
			}
		}
	}
}


func TestParserAddFlatten(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, true, "add", file)
		for i, _ := range paths {
			prs, err := Parse(json)
			if err != nil {
				t.Errorf("Total Fail(Parse Add), err:%v\n", err)
				return
			}
			err = prs.Add([]byte(`"test-value"`), paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Add), path:%v err:%v\n", paths[i], err)
				return
			}
			value, err := prs.Get(paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Add), path:%v err:%v\n", paths[i], err)
				return
			}
			if string(Flatten(value)) != values[i] {
				t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", paths[i], string(Flatten(value)), values[i])
				return
			}
		}
	}
}

func TestParserInsertFlatten(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, true, "insert", file)
		var value []byte
		prs, err := Parse(json)
		if err != nil {
			t.Errorf("Total Fail(Parse Insert), err:%v\n", err)
			return
		}
		for i, _ := range paths {
			empty, err := IsEmpty(json, paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Empty), path:%v err:%v\n", paths[i], err)
				return
			}
			if empty {
				t.Logf("SKIP(EMPTY ARRAY)")
				return
			}
			err = prs.Insert(0, []byte(`"test-value"`), paths[i]...)
			if err != nil {
				if err.Error() != EMPTY_ARRAY_ERROR().Error() {
					t.Errorf("Total Fail(Insert), path:%v err:%v\n", paths[i], err)
					return
				} else {
					continue
				}
			}
			value, err = prs.Get(paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Get Insert), path:%v err:%v\n", paths[i], err)
				return
			}
			if string(Flatten(value)) != stripQuotes(values[i]) {
				t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", paths[i], string(value), values[i])
				return
			}
		}
	}
}

func TestParserDeleteKVFlatten(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, true, "deleteKV", file)
		for i, _ := range paths {
			prs, err := Parse(json)
			if err != nil {
				t.Errorf("Total Fail(Parse Insert), err:%v\n", err)
				return
			}
			err = prs.AddKeyValue("test-key", []byte(`"test-value"`), paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Set), path:%v err:%v\n", paths[i], err)
				return
			}
			newPath := make([]string, len(paths[i]))
			copy(newPath, paths[i])
			newPath = append(newPath, "test-key")
			err = prs.Delete(newPath...)
			if err != nil {
				t.Errorf("Total Fail(Set), path:%v err:%v\n", paths[i], err)
				return
			}
			value, err := prs.Get(paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Get), path:%v err:%v\n", paths[i], err)
				return
			}
			if string(Flatten(value)) != stripQuotes(values[i]) {
				t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", paths[i], string(value), values[i])
				return
			}
		}
	}
}

func TestParserDeleteVFlatten(t *testing.T) {
	t.Logf("test files:")
	for f := 0 ; f < len(tests) ; f ++ {
		file := "tests" + tests[f]
		t.Logf("%v", file)
		InitValues(t, true, "deleteV", file)
		for i, _ := range paths {
			prs, err := Parse(json)
			if err != nil {
				t.Errorf("Total Fail(Parse Insert), err:%v\n", err)
				return
			}
			value, err := prs.Get(paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Get), path:%v err:%v\n", paths[i], err)
				return
			}
			arr := ParseArray(string(value))
			err = prs.Add([]byte(`"test-value"`), paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Set), path:%v err:%v\n", paths[i], err)
				return
			}
			newPath := make([]string, len(paths[i]))
			copy(newPath, paths[i])
			newPath = append(newPath, strconv.Itoa(len(arr)))
			err = prs.Delete(newPath...)
			if err != nil {
				t.Errorf("Total Fail(Set), path:%v err:%v\n", paths[i], err)
				return
			}
			value, err = prs.Get(paths[i]...)
			if err != nil {
				t.Errorf("Total Fail(Get), path:%v err:%v\n", paths[i], err)
				return
			}
			if string(Flatten(value)) != stripQuotes(values[i]) {
				t.Errorf("Fail, not same answer path:%v\n, got:\t\t>%v<\n, expected:\t>%v<\n", paths[i], string(value), values[i])
				return
			}
		}
	}
}

func TestParserEnd(t *testing.T) {
	InitValues(t, false, "get", "test/test-json.json")
}