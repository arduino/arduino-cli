package daemon

import "io"

func feedStream(streamer func(data []byte)) io.Writer {
	r, w := io.Pipe()
	go func() {
		data := make([]byte, 1024)
		for {
			if n, err := r.Read(data); err == nil {
				streamer(data[:n])
			} else {
				return
			}
		}
	}()
	return w
}
