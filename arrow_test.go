// Copyright (c) 2020-2022 Snowflake Computing Inc. All rights reserved.

package gosnowflake

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"
)

func TestArrowBigInt(t *testing.T) {
	db := openDB(t)
	dbt := &DBTest{t, db}

	testcases := []struct {
		num  string
		prec int
		sc   int
	}{
		{"10000000000000000000000000000000000000", 38, 0},
		{"-10000000000000000000000000000000000000", 38, 0},
		{"12345678901234567890123456789012345678", 38, 0},
		{"-12345678901234567890123456789012345678", 38, 0},
		{"99999999999999999999999999999999999999", 38, 0},
		{"-99999999999999999999999999999999999999", 38, 0},
	}

	for _, tc := range testcases {
		rows := dbt.mustQueryContext(WithHigherPrecision(context.Background()),
			fmt.Sprintf(selectNumberSQL, tc.num, tc.prec, tc.sc))
		if !rows.Next() {
			dbt.Error("failed to query")
		}
		defer rows.Close()
		var v *big.Int
		if err := rows.Scan(&v); err != nil {
			dbt.Errorf("failed to scan. %#v", err)
		}

		b, ok := new(big.Int).SetString(tc.num, 10)
		if !ok {
			dbt.Errorf("failed to convert %v big.Int.", tc.num)
		}
		if v.Cmp(b) != 0 {
			dbt.Errorf("big.Int value mismatch: expected %v, got %v", b, v)
		}
	}
}

func TestArrowBigFloat(t *testing.T) {
	db := openDB(t)
	dbt := &DBTest{t, db}

	testcases := []struct {
		num  string
		prec int
		sc   int
	}{
		{"1.23", 30, 2},
		{"1.0000000000000000000000000000000000000", 38, 37},
		{"-1.0000000000000000000000000000000000000", 38, 37},
		{"1.2345678901234567890123456789012345678", 38, 37},
		{"-1.2345678901234567890123456789012345678", 38, 37},
		{"9.9999999999999999999999999999999999999", 38, 37},
		{"-9.9999999999999999999999999999999999999", 38, 37},
	}

	for _, tc := range testcases {
		rows := dbt.mustQueryContext(WithHigherPrecision(context.Background()),
			fmt.Sprintf(selectNumberSQL, tc.num, tc.prec, tc.sc))
		if !rows.Next() {
			dbt.Error("failed to query")
		}
		defer rows.Close()
		var v *big.Float
		if err := rows.Scan(&v); err != nil {
			dbt.Errorf("failed to scan. %#v", err)
		}

		prec := v.Prec()
		b, ok := new(big.Float).SetPrec(prec).SetString(tc.num)
		if !ok {
			dbt.Errorf("failed to convert %v to big.Float.", tc.num)
		}
		if v.Cmp(b) != 0 {
			dbt.Errorf("big.Float value mismatch: expected %v, got %v", b, v)
		}
	}
}

func TestArrowIntPrecision(t *testing.T) {
	db := openDB(t)
	dbt := &DBTest{t, db}

	intTestcases := []struct {
		num  string
		prec int
		sc   int
	}{
		{"10000000000000000000000000000000000000", 38, 0},
		{"-10000000000000000000000000000000000000", 38, 0},
		{"12345678901234567890123456789012345678", 38, 0},
		{"-12345678901234567890123456789012345678", 38, 0},
		{"99999999999999999999999999999999999999", 38, 0},
		{"-99999999999999999999999999999999999999", 38, 0},
	}

	t.Run("arrow_disabled_scan_int64", func(t *testing.T) {
		for _, tc := range intTestcases {
			dbt.mustExec(forceJSON)
			rows := dbt.mustQuery(fmt.Sprintf(selectNumberSQL, tc.num, tc.prec, tc.sc))
			if !rows.Next() {
				dbt.Error("failed to query")
			}
			defer rows.Close()
			var v int64
			if err := rows.Scan(&v); err == nil {
				dbt.Error("should fail to scan")
			}
		}
	})
	t.Run("arrow_disabled_scan_string", func(t *testing.T) {
		for _, tc := range intTestcases {
			dbt.mustExec(forceJSON)
			rows := dbt.mustQuery(fmt.Sprintf(selectNumberSQL, tc.num, tc.prec, tc.sc))
			if !rows.Next() {
				dbt.Error("failed to query")
			}
			defer rows.Close()
			var v int64
			if err := rows.Scan(&v); err == nil {
				dbt.Error("should fail to scan")
			}
		}
	})
	t.Run("arrow_enabled_scan_big_int", func(t *testing.T) {
		for _, tc := range intTestcases {
			rows := dbt.mustQuery(fmt.Sprintf(selectNumberSQL, tc.num, tc.prec, tc.sc))
			if !rows.Next() {
				dbt.Error("failed to query")
			}
			defer rows.Close()
			var v string
			if err := rows.Scan(&v); err != nil {
				dbt.Errorf("failed to scan. %#v", err)
			}
			if !strings.EqualFold(v, tc.num) {
				dbt.Errorf("int value mismatch: expected %v, got %v", tc.num, v)
			}
		}
	})
	t.Run("arrow_high_precision_enabled_scan_big_int", func(t *testing.T) {
		for _, tc := range intTestcases {
			rows := dbt.mustQueryContext(
				WithHigherPrecision(context.Background()),
				fmt.Sprintf(selectNumberSQL, tc.num, tc.prec, tc.sc))
			if !rows.Next() {
				dbt.Error("failed to query")
			}
			defer rows.Close()
			var v *big.Int
			if err := rows.Scan(&v); err != nil {
				dbt.Errorf("failed to scan. %#v", err)
			}

			b, ok := new(big.Int).SetString(tc.num, 10)
			if !ok {
				dbt.Errorf("failed to convert %v big.Int.", tc.num)
			}
			if v.Cmp(b) != 0 {
				dbt.Errorf("big.Int value mismatch: expected %v, got %v", b, v)
			}
		}
	})
}

// TestArrowFloatPrecision tests the different variable types allowed in the
// rows.Scan() method. Note that for lower precision types we do not attempt
// to check the value as precision could be lost.
func TestArrowFloatPrecision(t *testing.T) {
	db := openDB(t)
	dbt := &DBTest{t, db}

	fltTestcases := []struct {
		num  string
		prec int
		sc   int
	}{
		{"1.23", 30, 2},
		{"1.0000000000000000000000000000000000000", 38, 37},
		{"-1.0000000000000000000000000000000000000", 38, 37},
		{"1.2345678901234567890123456789012345678", 38, 37},
		{"-1.2345678901234567890123456789012345678", 38, 37},
		{"9.9999999999999999999999999999999999999", 38, 37},
		{"-9.9999999999999999999999999999999999999", 38, 37},
	}

	t.Run("arrow_disabled_scan_float64", func(t *testing.T) {
		for _, tc := range fltTestcases {
			dbt.mustExec(forceJSON)
			rows := dbt.mustQuery(fmt.Sprintf(selectNumberSQL, tc.num, tc.prec, tc.sc))
			if !rows.Next() {
				dbt.Error("failed to query")
			}
			defer rows.Close()
			var v float64
			if err := rows.Scan(&v); err != nil {
				dbt.Errorf("failed to scan. %#v", err)
			}
		}
	})
	t.Run("arrow_disabled_scan_float32", func(t *testing.T) {
		for _, tc := range fltTestcases {
			dbt.mustExec(forceJSON)
			rows := dbt.mustQuery(fmt.Sprintf(selectNumberSQL, tc.num, tc.prec, tc.sc))
			if !rows.Next() {
				dbt.Error("failed to query")
			}
			defer rows.Close()
			var v float32
			if err := rows.Scan(&v); err != nil {
				dbt.Errorf("failed to scan. %#v", err)
			}
		}
	})
	t.Run("arrow_disabled_scan_string", func(t *testing.T) {
		for _, tc := range fltTestcases {
			dbt.mustExec(forceJSON)
			rows := dbt.mustQuery(fmt.Sprintf(selectNumberSQL, tc.num, tc.prec, tc.sc))
			if !rows.Next() {
				dbt.Error("failed to query")
			}
			defer rows.Close()
			var v string
			if err := rows.Scan(&v); err != nil {
				dbt.Errorf("failed to scan. %#v", err)
			}
			if !strings.EqualFold(v, tc.num) {
				dbt.Errorf("int value mismatch: expected %v, got %v", tc.num, v)
			}
		}
	})
	t.Run("arrow_enabled_scan_float64", func(t *testing.T) {
		for _, tc := range fltTestcases {
			rows := dbt.mustQuery(fmt.Sprintf(selectNumberSQL, tc.num, tc.prec, tc.sc))
			if !rows.Next() {
				dbt.Error("failed to query")
			}
			defer rows.Close()
			var v float64
			if err := rows.Scan(&v); err != nil {
				dbt.Errorf("failed to scan. %#v", err)
			}
		}
	})
	t.Run("arrow_enabled_scan_float32", func(t *testing.T) {
		for _, tc := range fltTestcases {
			rows := dbt.mustQuery(fmt.Sprintf(selectNumberSQL, tc.num, tc.prec, tc.sc))
			if !rows.Next() {
				dbt.Error("failed to query")
			}
			defer rows.Close()
			var v float32
			if err := rows.Scan(&v); err != nil {
				dbt.Errorf("failed to scan. %#v", err)
			}
		}
	})
	t.Run("arrow_enabled_scan_string", func(t *testing.T) {
		for _, tc := range fltTestcases {
			rows := dbt.mustQuery(fmt.Sprintf(selectNumberSQL, tc.num, tc.prec, tc.sc))
			if !rows.Next() {
				dbt.Error("failed to query")
			}
			defer rows.Close()
			var v string
			if err := rows.Scan(&v); err != nil {
				dbt.Errorf("failed to scan. %#v", err)
			}
		}
	})
	t.Run("arrow_high_precision_enabled_scan_big_float", func(t *testing.T) {
		for _, tc := range fltTestcases {
			rows := dbt.mustQueryContext(
				WithHigherPrecision(context.Background()),
				fmt.Sprintf(selectNumberSQL, tc.num, tc.prec, tc.sc))
			if !rows.Next() {
				dbt.Error("failed to query")
			}
			defer rows.Close()
			var v *big.Float
			if err := rows.Scan(&v); err != nil {
				dbt.Errorf("failed to scan. %#v", err)
			}

			prec := v.Prec()
			b, ok := new(big.Float).SetPrec(prec).SetString(tc.num)
			if !ok {
				dbt.Errorf("failed to convert %v to big.Float.", tc.num)
			}
			if v.Cmp(b) != 0 {
				dbt.Errorf("big.Float value mismatch: expected %v, got %v", b, v)
			}
		}
	})
}

func TestArrowVariousTypes(t *testing.T) {
	runTests(t, dsn, func(dbt *DBTest) {
		rows := dbt.mustQueryContext(
			WithHigherPrecision(context.Background()), selectVariousTypes)
		defer rows.Close()
		if !rows.Next() {
			dbt.Error("failed to query")
		}
		cc, err := rows.Columns()
		if err != nil {
			dbt.Errorf("columns: %v", cc)
		}
		ct, err := rows.ColumnTypes()
		if err != nil {
			dbt.Errorf("column types: %v", ct)
		}
		var v1 *big.Float
		var v2 int
		var v3 string
		var v4 float64
		var v5 []byte
		var v6 bool
		if err = rows.Scan(&v1, &v2, &v3, &v4, &v5, &v6); err != nil {
			dbt.Errorf("failed to scan: %#v", err)
		}
		if v1.Cmp(big.NewFloat(1.0)) != 0 {
			dbt.Errorf("failed to scan. %#v", *v1)
		}
		if ct[0].Name() != "C1" || ct[1].Name() != "C2" || ct[2].Name() != "C3" || ct[3].Name() != "C4" || ct[4].Name() != "C5" || ct[5].Name() != "C6" {
			dbt.Errorf("failed to get column names: %#v", ct)
		}
		if ct[0].ScanType() != reflect.TypeOf(float64(0)) {
			dbt.Errorf("failed to get scan type. expected: %v, got: %v", reflect.TypeOf(float64(0)), ct[0].ScanType())
		}
		if ct[1].ScanType() != reflect.TypeOf(int64(0)) {
			dbt.Errorf("failed to get scan type. expected: %v, got: %v", reflect.TypeOf(int64(0)), ct[1].ScanType())
		}
		var pr, sc int64
		var cLen int64
		pr, sc = dbt.mustDecimalSize(ct[0])
		if pr != 30 || sc != 2 {
			dbt.Errorf("failed to get precision and scale. %#v", ct[0])
		}
		dbt.mustFailLength(ct[0])
		if canNull := dbt.mustNullable(ct[0]); canNull {
			dbt.Errorf("failed to get nullable. %#v", ct[0])
		}
		if cLen != 0 {
			dbt.Errorf("failed to get length. %#v", ct[0])
		}
		if v2 != 2 {
			dbt.Errorf("failed to scan. %#v", v2)
		}
		pr, sc = dbt.mustDecimalSize(ct[1])
		if pr != 38 || sc != 0 {
			dbt.Errorf("failed to get precision and scale. %#v", ct[1])
		}
		dbt.mustFailLength(ct[1])
		if canNull := dbt.mustNullable(ct[1]); canNull {
			dbt.Errorf("failed to get nullable. %#v", ct[1])
		}
		if v3 != "t3" {
			dbt.Errorf("failed to scan. %#v", v3)
		}
		dbt.mustFailDecimalSize(ct[2])
		if cLen = dbt.mustLength(ct[2]); cLen != 2 {
			dbt.Errorf("failed to get length. %#v", ct[2])
		}
		if canNull := dbt.mustNullable(ct[2]); canNull {
			dbt.Errorf("failed to get nullable. %#v", ct[2])
		}
		if v4 != 4.2 {
			dbt.Errorf("failed to scan. %#v", v4)
		}
		dbt.mustFailDecimalSize(ct[3])
		dbt.mustFailLength(ct[3])
		if canNull := dbt.mustNullable(ct[3]); canNull {
			dbt.Errorf("failed to get nullable. %#v", ct[3])
		}
		if !bytes.Equal(v5, []byte{0xab, 0xcd}) {
			dbt.Errorf("failed to scan. %#v", v5)
		}
		dbt.mustFailDecimalSize(ct[4])
		if cLen = dbt.mustLength(ct[4]); cLen != 8388608 { // BINARY
			dbt.Errorf("failed to get length. %#v", ct[4])
		}
		if canNull := dbt.mustNullable(ct[4]); canNull {
			dbt.Errorf("failed to get nullable. %#v", ct[4])
		}
		if !v6 {
			dbt.Errorf("failed to scan. %#v", v6)
		}
		dbt.mustFailDecimalSize(ct[5])
		dbt.mustFailLength(ct[5])
		/*canNull = dbt.mustNullable(ct[5])
		if canNull {
			dbt.Errorf("failed to get nullable. %#v", ct[5])
		}*/
	})
}
