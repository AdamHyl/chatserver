package utility

import (
	"fmt"
	"github.com/rcrowley/go-bson"
)

func GetUID() string {
	return fmt.Sprintf("%x", string(bson.NewObjectId()))
}
