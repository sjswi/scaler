package main

import (
	"fmt"
	"strings"

	"vitess.io/vitess/go/vt/sqlparser"
)

func extractTableNames(sql string) ([]string, error) {
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return nil, err
	}

	var tables []string
	sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
		switch node := node.(type) {
		case *sqlparser.AliasedTableExpr:
			if tableName, ok := node.Expr.(sqlparser.TableName); ok {
				tables = append(tables, tableName.Name.String())
			}
		case *sqlparser.CreateTable:

			tables = append(tables, node.Table.Name.String())

		}

		return true, nil
	}, stmt)

	return tables, nil
}

func main() {
	sql := " CREATE TABLE orders (id INT);"
	tables, err := extractTableNames(sql)
	if err != nil {
		fmt.Println("Error parsing SQL:", err)
		return
	}

	fmt.Println("Tables:", strings.Join(tables, ", "))
}
