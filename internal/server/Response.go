package server

import (
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
	kratosstatus "github.com/go-kratos/kratos/v2/transport/http/status"
	"google.golang.org/grpc/status"
	"net/http"
)

// 返回正确
func SuccessResponse(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if v == nil {
		return nil
	}
	if rd, ok := v.(kratoshttp.Redirector); ok {
		url, code := rd.Redirect()
		http.Redirect(w, r, url, code)
		return nil
	}
	var susresp Response
	susresp = Response{
		Code:    200,
		Message: "success",
		Data:    v,
	}
	codec, _ := kratoshttp.CodecForRequest(r, "Accept")
	data, err := codec.Marshal(susresp)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/"+(codec.Name()))
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// 返回错误码
func ErrorResponseEncoder(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}
	var errresp Response
	var se *status.Status
	var ok bool
	if se, ok = status.FromError(err); ok {
		errresp = Response{
			Code:    kratosstatus.FromGRPCCode(se.Code()),
			Message: se.Message(),
			Data:    nil,
		}
	} else {
		errresp = Response{
			Code:    500,
			Message: "internal logic has error",
			Data:    nil,
		}
	}
	if rd, ok := err.(kratoshttp.Redirector); ok {
		url, code := rd.Redirect()
		http.Redirect(w, r, url, code)
		return
	}
	codec, _ := kratoshttp.CodecForRequest(r, "Accept")
	data, err := codec.Marshal(errresp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/"+codec.Name())
	w.WriteHeader(errresp.Code)
	_, _ = w.Write(data)
	return
}
