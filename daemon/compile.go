//

package daemon

/*
import (
	"time"

	pb "github.com/arduino/arduino-cli/commands"
)

func (d *daemon) Compile(req *pb.CompileReq, serv pb.ArduinoCore_CompileServer) error {
	if err := serv.Send(&pb.CompileResp{Result: "OK"}); err != nil {
		return err
	}
	for i := 0; i < 3; i++ {
		time.Sleep(time.Second)
		if err := serv.Send(&pb.CompileResp{OutputLine: "a\n"}); err != nil {
			return err
		}
	}
	return nil
}
*/
