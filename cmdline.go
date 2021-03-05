package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ReadInput(message string) string {
	if message != "" {
		fmt.Printf("%s: ", message)
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	return scanner.Text()
}

func ReadInputD(message string, d string) string {
	m := ReadInput(message)
	if m == "" {
		return d
	}

	return m
}

func ReadInputAsArray(message string, sep string) []string {
	return strings.Split(ReadInput(message+" [seperator: "+sep+"] "), sep)
}

func ReadInputAsArrayD(message string, sep string, d []string) []string {
	m := ReadInputAsArray(message, sep)
	if len(m) == 0 {
		return d
	}

	return m
}

func ReadInputAsBool(message string, cond string) bool {
	res := ReadInput(message + "[" + cond + "] ")
	if res == cond {
		return true
	}

	return false
}

func ReadInputAsInt(message string) (int, error) {
	var i int
	var err error

	n := ReadInput(message)
	if n == "" {
		return 0, nil
	}

	i, err = strconv.Atoi(n)
	if checkError(err) {
		return 0, err
	}

	return i, nil
}

func ReadInputAsIntD(message string, d int) (int, error) {
	m, err := ReadInputAsInt(message)
	if checkError(err) {
		return 0, err
	}

	if m == 0 {
		return d, nil
	}

	return m, nil
}

func ReadInputAsFloat(message string) (float32, error) {
	var f float64
	var err error

	n := ReadInput(message)
	if n == "" {
		return 0, nil
	}

	f, err = strconv.ParseFloat(n, 32)
	if checkError(err) {
		return 0, err
	}

	return float32(f), nil
}

func ReadInputAsFloatD(message string, d float32) (float32, error) {
	m, err := ReadInputAsFloat(message)
	if checkError(err) {
		return 0, err
	}

	if m == 0 {
		return d, nil
	}

	return m, nil
}
