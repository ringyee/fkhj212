package storage

import (
	//"context"
	"github.com/jmoiron/sqlx"
	"time"
	// register postgresql driver
	_ "github.com/lib/pq"
	//"github.com/pkg/errors"
	//log "github.com/sirupsen/logrus"
)

// Factors defines a LoRaWAN device.
type Factors struct {
	ValueTimeStamp time.Time `db:"value_timestamp"`
	DeviceID       string    `db:"device_id"`
	Value1         float64   `db:"value_1"`
	Value2         float64   `db:"value_2"`
	Value3         float64   `db:"value_3"`
	Value4         float64   `db:"value_4"`
	Value5         float64   `db:"value_5"`
	Value6         float64   `db:"value_6"`
	Value7         float64   `db:"value_7"`
	Value8         float64   `db:"value_8"`
	Value9         float64   `db:"value_9"`
	Value10        float64   `db:"value_10"`
}

//InsertFactorValue ... strore Factors value
func InsertFactorValue(db sqlx.Ext, d *Factors) error {
	_, err := db.Exec(`
	insert into device_value (
	value_timestamp,
	device_id,
	value_1,
	value_2,
	value_3,
	value_4,
	value_5,
	value_6,
	value_7,
	value_8,
	value_9,
	value_10
	) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,$12)`,
		d.ValueTimeStamp,
		d.DeviceID,
		d.Value1,
		d.Value2,
		d.Value3,
		d.Value4,
		d.Value5,
		d.Value6,
		d.Value7,
		d.Value8,
		d.Value9,
		d.Value10,
	)
	if err != nil {
		return err
	}
	return nil
}
