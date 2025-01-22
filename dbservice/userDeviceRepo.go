package dbservice

import (
	"log"

	"github.com/scylladb/gocqlx/table"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/qb"
)

type UserDeviceRepository struct {
	session *gocqlx.Session
}


type UserDevice struct {
	UserId string `json:"user_id"`
	DeviceId string `json:"device_id"`
	RedisId string `json:"redis_id"`
}


var UserDeviceMetadata = table.Metadata{
	Name: "user_device",
	Columns: []string{"user_id", "device_id", "redis_id"},
	PartKey: []string{"device_id"},
}

var UserDeviceTable = table.New(UserDeviceMetadata)

func NewUserDeviceRepository(session *gocqlx.Session) *UserDeviceRepository {	

	return &UserDeviceRepository{
		session: session,
	}
}


func (u *UserDeviceRepository) AddUserDevice(user_id string, device_id string) error {

	// checking if device exists
	
	insertQuery := u.session.Query(UserDeviceTable.Insert()).BindStruct(&UserDevice{UserId: user_id, DeviceId: device_id, RedisId: ""})
	
	if err := insertQuery.ExecRelease(); err != nil {
		log.Println("error while inserting user device : ", err)
		return err
	}

	return nil
}

func (u *UserDeviceRepository) UpdateUserDevice(device map[string]interface{},user_id string) error {

	updataQuery := qb.Update("store.user_device")
    qb_map := qb.M{}
	for key, value := range device {
		updataQuery.Set(key)
		qb_map[key] = value
	}

	updataQuery.Where(qb.Eq("device_id"))
	qb_map["device_id"] = user_id

	stmt, names := updataQuery.ToCql()

	if err := u.session.Query(stmt, names).BindMap(qb_map).ExecRelease(); err != nil {
		log.Println("error while updating user device : ", err)
		return err
	}

	return nil
}


func (u *UserDeviceRepository) GetUserDevice(device_id string) (*UserDevice, error) {


	userDevice := &UserDevice{}
	query := u.session.Query(UserDeviceTable.Get()).BindMap(map[string]interface{}{"device_id": device_id})
	if err := query.GetRelease(userDevice); err != nil {
		log.Println("error while fetching user device : ", err)
		return nil, err
	}
	return userDevice, nil
}


func(u *UserDeviceRepository) LastMsgIdRead(user_id string) *UserDevice {
	userDevice , err := u.GetUserDevice(user_id)
	if err != nil {
		log.Println("error while fetching user device : ", err)
		return &UserDevice{}
	}
	return userDevice
}

func (u *UserDeviceRepository) DeleteUserDevice(device_id string) error {
	err := u.session.Query(UserDeviceTable.Delete()).BindMap(map[string]interface{}{"device_id": device_id}).ExecRelease()
	if err != nil {
		log.Println("error while deleting user device : ", err)
		return err
	}
	return nil
}


func (u *UserDeviceRepository) checkIfDeviceExists(device_id string) (bool , error){
	stmt, names := qb.Select("store.user_device").Count("device_id ").Where(qb.Eq("device_id")).ToCql()
	res := 0
	err := u.session.Query(stmt, names).BindMap(qb.M{"device_id": device_id}).Get(&res)
	if err != nil {
		log.Println("error while checking if device exists : ", err)
		return false, err
	}

	if res > 0 {
		return true, nil
	}else {
		return false, nil
	}
}
