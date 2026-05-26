package repository

import (
	"context"

	"github.com/ronexlemon/bnbcore/internal/domain/unit"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/db"
)


type UnitRepository struct{
	DbConnection *db.PostgresConn
}

func NewUnitRepository(dbconn *db.PostgresConn)*UnitRepository{
	return&UnitRepository{DbConnection: dbconn}
}

func (u *UnitRepository)Create(ctx context.Context, unit *unit.Unit) (*unit.Unit, error){

	query:=`INSERT INTO units(id,tenant_id,title,description,price_per_night,location,latitude,longitude,status)
	VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`

	_,err:=u.DbConnection.Pool.Exec(ctx,query,unit.ID,unit.TenantID,unit.Title,unit.Description,unit.PricePerNight,unit.Location,unit.Latitude,unit.Longitude,unit.Status)
	return unit,err
}