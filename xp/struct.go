package xp

type Person struct {
	BigIntCol     int64   `parquet:"name=bigint_col"`
	BoolCol       bool    `parquet:"name=bool_col"`
	DateStringCol string  `parquet:"name=date_string_col"`
	DoubleCol     float64 `parquet:"name=double_col"`
	FloatCol      float32 `parquet:"name=float_col"`
	Id            int32   `parquet:"name=id"`
	IntCol        int32   `parquet:"name=int_col"`
	SmallintCol   int32   `parquet:"name=smallint_col"`
	StringCol     string  `parquet:"name=string_col"`
	TimestampCol  int64   `parquet:"name=timestamp_col"`
	TinyintCol    int32   `parquet:"name=tinyint_col"`
}

type Student struct {
	Name    string           `parquet:"name=name, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Age     int32            `parquet:"name=age, type=INT32"`
	Id      int64            `parquet:"name=id, type=INT64"`
	Weight  float32          `parquet:"name=weight, type=FLOAT"`
	Sex     bool             `parquet:"name=sex, type=BOOLEAN"`
	Day     int32            `parquet:"name=day, type=INT32, convertedtype=DATE"`
	Scores  map[string]int32 `parquet:"name=scores, type=MAP, keytype=BYTE_ARRAY, keyconvertedtype=UTF8, valuetype=INT32"`
	Ignored int32            //without parquet tag and won't write
}
