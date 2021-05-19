// Copyright (C) 2021, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package string

import (
	"testing"

	asserts "github.com/stretchr/testify/assert"
)

// Test_stringSliceContainsString tests the SliceContainsString function
func Test_stringSliceContainsString(t *testing.T) {
	assert := asserts.New(t)
	var slice []string
	var find string
	var found bool

	// GIVEN a nil slice
	// WHEN an empty string is searched for
	// THEN verify false is returned
	slice = nil
	found = SliceContainsString(slice, find)
	assert.Equal(found, false)

	// GIVEN a slice with several strings
	// WHEN one of the strings is searched for
	// THEN verify string is found
	slice = []string{"test-value-1", "test-value-2", "test-value-3"}
	find = "test-value-2"
	found = SliceContainsString(slice, find)
	assert.Equal(found, true)

	// GIVEN a slice with several strings
	// WHEN a string not in the slice is searched for
	// THEN verify string is not found
	slice = []string{"test-value-1", "test-value-2", "test-value-3"}
	find = "test-value-4"
	found = SliceContainsString(slice, find)
	assert.Equal(found, false)
}

// Test_stringSliceContainsString tests the RemoveStringFromSlice function
func Test_removeStringFromStringSlice(t *testing.T) {
	assert := asserts.New(t)
	var slice []string
	var remove string
	var output []string

	// GIVEN a nil slice and an empty string to remove
	// WHEN the empty string is removed from the nil slice
	// THEN verify that an empty slice is returned
	slice = nil
	remove = ""
	output = RemoveStringFromSlice(slice, remove)
	assert.NotNil(output)
	assert.Len(output, 0)

	// GIVEN a slice with several strings
	// WHEN a string in the slice is removed
	// THEN verify slice is correct
	slice = []string{"test-value-1", "test-value-2", "test-value-3"}
	remove = "test-value-2"
	output = RemoveStringFromSlice(slice, remove)
	assert.Equal("test-value-1", slice[0])
	assert.Equal("test-value-2", slice[1])
	assert.Len(output, 2)
}

// TestUnorderedEqual tests the UnorderedEqual function
func TestUnorderedEqual(t *testing.T) {
	assert := asserts.New(t)
	var mapBool map[string]bool
	var arrayStr []string

	// GIVEN a map and array with the same elements and order
	// WHEN compared
	// THEN the UnorderedEqual returns true
	arrayStr = []string{"test-value-1", "test-value-2", "test-value-3"}
	mapBool = make(map[string]bool)
	mapBool["test-value-1"] = true
	mapBool["test-value-2"] = true
	mapBool["test-value-3"] = true
	success := UnorderedEqual(mapBool, arrayStr)
	assert.Equal(true, success)

	// GIVEN a map and array with the same elements and different order
	// WHEN compared
	// THEN the UnorderedEqual returns true
	arrayStr = []string{"test-value-2", "test-value-3", "test-value-1"}
	mapBool = make(map[string]bool)
	mapBool["test-value-1"] = true
	mapBool["test-value-2"] = true
	mapBool["test-value-3"] = true
	success = UnorderedEqual(mapBool, arrayStr)
	assert.Equal(true, success)

	// GIVEN a map and array with the different number of elements
	// WHEN compared
	// THEN the UnorderedEqual returns false
	arrayStr = []string{"test-value-2", "test-value-3"}
	mapBool = make(map[string]bool)
	mapBool["test-value-1"] = true
	mapBool["test-value-2"] = true
	mapBool["test-value-3"] = true
	success = UnorderedEqual(mapBool, arrayStr)
	assert.Equal(false, success)

	// GIVEN a map and array with the same number of elements but different elements
	// WHEN compared
	// THEN the UnorderedEqual returns false
	arrayStr = []string{"test-value-2", "test-value-3", "test-value-4"}
	mapBool = make(map[string]bool)
	mapBool["test-value-1"] = true
	mapBool["test-value-5"] = true
	mapBool["test-value-3"] = true
	success = UnorderedEqual(mapBool, arrayStr)
	assert.Equal(false, success)
}

// TestSliceToSet tests the SliceContainsString function
func TestSliceToSet(t *testing.T) {
	assert := asserts.New(t)
	slice := []string{"s1", "s2", "s3"}

	// GIVEN a slice with several strings
	// WHEN the slice is converted to a set
	// THEN verify the set is correct
	set := SliceToSet(slice)
	assert.Len(set, 3)
	assert.Contains(slice, "s1", "Set should contain string")
	assert.Contains(slice, "s2", "Set should contain string")
	assert.Contains(slice, "s3", "Set should contain string")
	assert.NotContains(slice, "s4", "Set should not contain string")
}

// TestEmptyOrNilSliceToSet tests the SliceContainsString function
func TestEmptyOrNilSliceToSet(t *testing.T) {
	assert := asserts.New(t)
	slice := []string{}

	// GIVEN an empty slice
	// WHEN the slice is converted to a set
	// THEN verify the set is empty
	set := SliceToSet(slice)
	assert.Len(set, 0, "Empty slice should result in empty set")

	// GIVEN an nil slice
	// WHEN the slice is converted to a set
	// THEN verify the set is empty
	set = SliceToSet(nil)
	assert.Len(set, 0, "Nil slice should result in empty set")
}
