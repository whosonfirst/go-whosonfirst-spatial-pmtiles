package request

import (
	"github.com/whosonfirst/go-whosonfirst-spatial-grpc/spatial"
)

func existentialIntFlagsToProtobufExistentialFlags(int_flags []int64) []spatial.ExistentialFlag {

	sp_flags := make([]spatial.ExistentialFlag, len(int_flags))

	for idx, i := range int_flags {

		var fl spatial.ExistentialFlag

		switch i {
		case 0:
			fl = spatial.ExistentialFlag_FALSE
		case 1:
			fl = spatial.ExistentialFlag_TRUE
		default:
			fl = spatial.ExistentialFlag_UNKNOWN
		}

		sp_flags[idx] = fl
	}

	return sp_flags
}

func protobufExistentalFlagsToExistentialIntFlags(sp_flags []spatial.ExistentialFlag) []int64 {

	int_flags := make([]int64, len(sp_flags))

	for idx, fl := range sp_flags {

		var i int64

		switch fl {
		case spatial.ExistentialFlag_FALSE:
			i = 0
		case spatial.ExistentialFlag_TRUE:
			i = 1
		default:
			i = -1
		}

		int_flags[idx] = i
	}

	return int_flags
}
