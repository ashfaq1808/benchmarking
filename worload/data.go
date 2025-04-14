package workload

import (
	"math/rand"
	"strconv"
	"time"
)

type Employee struct {
	ID     string
	Name   string
	Dept   string
	Salary int
}

var names = []string{"Alice", "Bob", "Charlie", "Diana", "Eve"}
var depts = []string{"HR", "Engineering", "Finance", "Sales", "Legal"}

func GenerateEmployee(id int) Employee {
	rand.Seed(time.Now().UnixNano())
	return Employee{
		ID:     strconv.Itoa(id),
		Name:   names[rand.Intn(len(names))],
		Dept:   depts[rand.Intn(len(depts))],
		Salary: rand.Intn(90000) + 30000,
	}
}
