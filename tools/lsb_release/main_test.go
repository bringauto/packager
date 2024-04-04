package main

import "testing"

func lsb_release_test(distId string, relNumber string, validate bool, file string, t *testing.T) {
	data := DataStruct{}

	data.ReadFromFile(file, validate)
	if data.DistributorID != distId {
		t.Errorf("Distributor ID is not %s", distId)
	}
	if data.ReleaseNumber != relNumber {
		t.Errorf("Release number is not %s", relNumber)
	}
}

func Test_lsb_release_multipledigit(t *testing.T) {
	lsb_release_test("Ubuntu", "18.04", false, "test_data/lsb_release_ubuntu1804.txt", t)
}

func Test_lsb_release_digit(t *testing.T) {
	lsb_release_test("Debian", "11", false, "test_data/lsb_release.txt", t)
}

func Test_lsb_release_validate_ok(t *testing.T) {
	lsb_release_test("Debian", "11", true, "test_data/lsb_release.txt", t)
	lsb_release_test("Ubuntu", "18.04", true, "test_data/lsb_release_ubuntu1804.txt", t)
}

func Test_lsb_release_validate_notok(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	lsb_release_test("Debian", "11", true, "test_data/lsb_release_broken.txt", t)
}

func Test_lsb_release_validate_notok_nopanic(t *testing.T) {
	data := DataStruct{}
	data.ReadFromFile("test_data/lsb_release_broken.txt", false)
}
