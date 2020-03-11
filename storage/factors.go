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

//// GetFactors returns the device matching the given DevEUI.
//func GetFactors(db sqlx.Queryer, devEUI lorawan.EUI64) (Factors, error) {
//var d Factors
//err := sqlx.Get(db, &d, "select * from device where dev_eui = $1", devEUI[:])
//if err != nil {
//return d, handlePSQLError(Select, err, "select error")
//}

//n, err := GetNetworkServerForDevEUI(db, d.DevEUI)
//if err != nil {
//return d, errors.Wrap(err, "get network-server error")
//}

//nsClient, err := config.C.NetworkServer.Pool.Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
//if err != nil {
//return d, errors.Wrap(err, "get network-server client error")
//}

//resp, err := nsClient.GetFactors(context.Background(), &ns.GetFactorsRequest{
//DevEUI: d.DevEUI[:],
//})
//if err != nil {
//return d, err
//}

//if resp.Factors != nil {
//d.SkipFCntCheck = resp.Factors.SkipFCntCheck
//}

//return d, nil
//}

//// UpdateFactors updates the given device.
//func UpdateFactors(db sqlx.Ext, d *Factors) error {
//if err := d.Validate(); err != nil {
//return errors.Wrap(err, "validate error")
//}

//d.UpdatedAt = time.Now()

//res, err := db.Exec(`
//update device
//set
//updated_at = $2,
//application_id = $3,
//device_profile_id = $4,
//name = $5,
//description = $6,
//device_status_battery = $7,
//device_status_margin = $8,
//last_seen_at = $9
//where
//dev_eui = $1`,
//d.DevEUI[:],
//d.UpdatedAt,
//d.ApplicationID,
//d.FactorsProfileID,
//d.Name,
//d.Description,
//d.FactorsStatusBattery,
//d.FactorsStatusMargin,
//d.LastSeenAt,
//)
//if err != nil {
//return handlePSQLError(Update, err, "update error")
//}
//ra, err := res.RowsAffected()
//if err != nil {
//return errors.Wrap(err, "get rows affected error")
//}
//if ra == 0 {
//return ErrDoesNotExist
//}

//app, err := GetApplication(db, d.ApplicationID)
//if err != nil {
//return errors.Wrap(err, "get application error")
//}

//n, err := GetNetworkServerForDevEUI(db, d.DevEUI)
//if err != nil {
//return errors.Wrap(err, "get network-server error")
//}

//nsClient, err := config.C.NetworkServer.Pool.Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
//if err != nil {
//return errors.Wrap(err, "get network-server client error")
//}

//_, err = nsClient.UpdateFactors(context.Background(), &ns.UpdateFactorsRequest{
//Factors: &ns.Factors{
//DevEUI:           d.DevEUI[:],
//FactorsProfileID: d.FactorsProfileID,
//ServiceProfileID: app.ServiceProfileID,
//RoutingProfileID: config.C.ApplicationServer.ID,
//SkipFCntCheck:    d.SkipFCntCheck,
//},
//})
//if err != nil {
//log.WithError(err).Error("network-server update device api error")
//return handleGrpcError(err, "update device error")
//}

//log.WithFields(log.Fields{
//"dev_eui": d.DevEUI,
//}).Info("device updated")

//return nil
//}

//// DeleteFactors deletes the device matching the given DevEUI.
//func DeleteFactors(db sqlx.Ext, devEUI lorawan.EUI64) error {
//n, err := GetNetworkServerForDevEUI(db, devEUI)
//if err != nil {
//return errors.Wrap(err, "get network-server error")
//}

//res, err := db.Exec("delete from device where dev_eui = $1", devEUI[:])
//if err != nil {
//return handlePSQLError(Delete, err, "delete error")
//}
//ra, err := res.RowsAffected()
//if err != nil {
//return errors.Wrap(err, "get rows affected error")
//}
//if ra == 0 {
//return ErrDoesNotExist
//}

//nsClient, err := config.C.NetworkServer.Pool.Get(n.Server, []byte(n.CACert), []byte(n.TLSCert), []byte(n.TLSKey))
//if err != nil {
//return errors.Wrap(err, "get network-server client error")
//}

//_, err = nsClient.DeleteFactors(context.Background(), &ns.DeleteFactorsRequest{
//DevEUI: devEUI[:],
//})
//if err != nil && grpc.Code(err) != codes.NotFound {
//log.WithError(err).Error("network-server delete device api error")
//return handleGrpcError(err, "delete device error")
//}

//log.WithFields(log.Fields{
//"dev_eui": devEUI,
//}).Info("device deleted")

//return nil
//}
