package pkg

import (
	sf "github.com/bwmarrin/snowflake"
	"time"
)

var SnowFlake = new(sf.Node)

func Init(start_time string, machine_id int64) error {
	var str time.Time
	str, err := time.Parse("2012-12-21", start_time)
	if err != nil {
		return err
	}
	sf.Epoch = str.UnixNano()
	SnowFlake, err = sf.NewNode(machine_id)
	if err != nil {
		return err
	}
	return nil
}

func Gen() int64 {
	return SnowFlake.Generate().Int64()
}
