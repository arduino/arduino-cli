# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [monitor/monitor.proto](#monitor/monitor.proto)
    - [MonitorConfig](#cc.arduino.cli.monitor.MonitorConfig)
    - [StreamingOpenReq](#cc.arduino.cli.monitor.StreamingOpenReq)
    - [StreamingOpenResp](#cc.arduino.cli.monitor.StreamingOpenResp)
  
    - [MonitorConfig.TargetType](#cc.arduino.cli.monitor.MonitorConfig.TargetType)
  
  
    - [Monitor](#cc.arduino.cli.monitor.Monitor)
  

- [Scalar Value Types](#scalar-value-types)



<a name="monitor/monitor.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## monitor/monitor.proto



<a name="cc.arduino.cli.monitor.MonitorConfig"></a>

### MonitorConfig
Tells the monitor which target to open and provides additional parameters
that might be needed to configure the target or the monitor itself.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| target | [string](#string) |  |  |
| type | [MonitorConfig.TargetType](#cc.arduino.cli.monitor.MonitorConfig.TargetType) |  |  |
| additionalConfig | [google.protobuf.Struct](#google.protobuf.Struct) |  |  |






<a name="cc.arduino.cli.monitor.StreamingOpenReq"></a>

### StreamingOpenReq
The top-level message sent by the client for the `StreamingOpen` method.
Multiple `StreamingOpenReq` messages can be sent but the first message
must contain a `monitor_config` message to initialize the monitor target.
All subsequent messages must contain bytes to be sent to the target
and must not contain a `monitor_config` message.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| monitorConfig | [MonitorConfig](#cc.arduino.cli.monitor.MonitorConfig) |  | Provides information to the monitor that specifies which is the target. The first `StreamingOpenReq` message must contain a `monitor_config` message. |
| data | [bytes](#bytes) |  | The data to be sent to the target being monitored. |






<a name="cc.arduino.cli.monitor.StreamingOpenResp"></a>

### StreamingOpenResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| data | [bytes](#bytes) |  |  |





 


<a name="cc.arduino.cli.monitor.MonitorConfig.TargetType"></a>

### MonitorConfig.TargetType


| Name | Number | Description |
| ---- | ------ | ----------- |
| SERIAL | 0 |  |


 

 


<a name="cc.arduino.cli.monitor.Monitor"></a>

### Monitor
Service that abstract a Monitor usage

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| StreamingOpen | [StreamingOpenReq](#cc.arduino.cli.monitor.StreamingOpenReq) stream | [StreamingOpenResp](#cc.arduino.cli.monitor.StreamingOpenResp) stream |  |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

