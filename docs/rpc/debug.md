# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [debug/debug.proto](#debug/debug.proto)
    - [DebugConfigReq](#cc.arduino.cli.debug.DebugConfigReq)
    - [DebugReq](#cc.arduino.cli.debug.DebugReq)
    - [DebugResp](#cc.arduino.cli.debug.DebugResp)
    - [Instance](#cc.arduino.cli.debug.Instance)
  
  
  
    - [Debug](#cc.arduino.cli.debug.Debug)
  

- [Scalar Value Types](#scalar-value-types)



<a name="debug/debug.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## debug/debug.proto



<a name="cc.arduino.cli.debug.DebugConfigReq"></a>

### DebugConfigReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.debug.Instance) |  |  |
| fqbn | [string](#string) |  |  |
| sketch_path | [string](#string) |  |  |
| port | [string](#string) |  |  |
| verbose | [bool](#bool) |  |  |
| import_file | [string](#string) |  |  |






<a name="cc.arduino.cli.debug.DebugReq"></a>

### DebugReq
The top-level message sent by the client for the `Debug` method.
Multiple `DebugReq` messages can be sent but the first message
must contain a `DebugReq` message to initialize the debug session.
All subsequent messages must contain bytes to be sent to the debug session
and must not contain a `DebugReq` message.

Content must be either a debug session config or data to be sent.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| debugReq | [DebugConfigReq](#cc.arduino.cli.debug.DebugConfigReq) |  | Provides information to the debug that specifies which is the target. The first `StreamingOpenReq` message must contain a `DebugReq` message. |
| data | [bytes](#bytes) |  | The data to be sent to the target being monitored. |
| send_interrupt | [bool](#bool) |  | Set this to true to send and Interrupt signal to the debugger process |






<a name="cc.arduino.cli.debug.DebugResp"></a>

### DebugResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| data | [bytes](#bytes) |  |  |
| error | [string](#string) |  |  |






<a name="cc.arduino.cli.debug.Instance"></a>

### Instance
TODO remove this in next proto refactoring because is a duplicate from commands/common.proto


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int32](#int32) |  |  |





 

 

 


<a name="cc.arduino.cli.debug.Debug"></a>

### Debug
Service that abstract a debug Session usage

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Debug | [DebugReq](#cc.arduino.cli.debug.DebugReq) stream | [DebugResp](#cc.arduino.cli.debug.DebugResp) stream |  |

 



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

