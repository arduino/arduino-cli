# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [commands/board.proto](#commands/board.proto)
    - [BoardAttachReq](#cc.arduino.cli.commands.BoardAttachReq)
    - [BoardAttachResp](#cc.arduino.cli.commands.BoardAttachResp)
    - [BoardDetailsReq](#cc.arduino.cli.commands.BoardDetailsReq)
    - [BoardDetailsResp](#cc.arduino.cli.commands.BoardDetailsResp)
    - [BoardListAllReq](#cc.arduino.cli.commands.BoardListAllReq)
    - [BoardListAllResp](#cc.arduino.cli.commands.BoardListAllResp)
    - [BoardListItem](#cc.arduino.cli.commands.BoardListItem)
    - [BoardListReq](#cc.arduino.cli.commands.BoardListReq)
    - [BoardListResp](#cc.arduino.cli.commands.BoardListResp)
    - [ConfigOption](#cc.arduino.cli.commands.ConfigOption)
    - [ConfigValue](#cc.arduino.cli.commands.ConfigValue)
    - [DetectedPort](#cc.arduino.cli.commands.DetectedPort)
    - [RequiredTool](#cc.arduino.cli.commands.RequiredTool)
  
  
  
  

- [commands/commands.proto](#commands/commands.proto)
    - [DestroyReq](#cc.arduino.cli.commands.DestroyReq)
    - [DestroyResp](#cc.arduino.cli.commands.DestroyResp)
    - [InitReq](#cc.arduino.cli.commands.InitReq)
    - [InitResp](#cc.arduino.cli.commands.InitResp)
    - [RescanReq](#cc.arduino.cli.commands.RescanReq)
    - [RescanResp](#cc.arduino.cli.commands.RescanResp)
    - [UpdateIndexReq](#cc.arduino.cli.commands.UpdateIndexReq)
    - [UpdateIndexResp](#cc.arduino.cli.commands.UpdateIndexResp)
    - [UpdateLibrariesIndexReq](#cc.arduino.cli.commands.UpdateLibrariesIndexReq)
    - [UpdateLibrariesIndexResp](#cc.arduino.cli.commands.UpdateLibrariesIndexResp)
    - [VersionReq](#cc.arduino.cli.commands.VersionReq)
    - [VersionResp](#cc.arduino.cli.commands.VersionResp)
  
  
  
    - [ArduinoCore](#cc.arduino.cli.commands.ArduinoCore)
  

- [commands/common.proto](#commands/common.proto)
    - [DownloadProgress](#cc.arduino.cli.commands.DownloadProgress)
    - [Instance](#cc.arduino.cli.commands.Instance)
    - [TaskProgress](#cc.arduino.cli.commands.TaskProgress)
  
  
  
  

- [commands/compile.proto](#commands/compile.proto)
    - [CompileReq](#cc.arduino.cli.commands.CompileReq)
    - [CompileResp](#cc.arduino.cli.commands.CompileResp)
  
  
  
  

- [commands/core.proto](#commands/core.proto)
    - [Board](#cc.arduino.cli.commands.Board)
    - [Platform](#cc.arduino.cli.commands.Platform)
    - [PlatformDownloadReq](#cc.arduino.cli.commands.PlatformDownloadReq)
    - [PlatformDownloadResp](#cc.arduino.cli.commands.PlatformDownloadResp)
    - [PlatformInstallReq](#cc.arduino.cli.commands.PlatformInstallReq)
    - [PlatformInstallResp](#cc.arduino.cli.commands.PlatformInstallResp)
    - [PlatformListReq](#cc.arduino.cli.commands.PlatformListReq)
    - [PlatformListResp](#cc.arduino.cli.commands.PlatformListResp)
    - [PlatformSearchReq](#cc.arduino.cli.commands.PlatformSearchReq)
    - [PlatformSearchResp](#cc.arduino.cli.commands.PlatformSearchResp)
    - [PlatformUninstallReq](#cc.arduino.cli.commands.PlatformUninstallReq)
    - [PlatformUninstallResp](#cc.arduino.cli.commands.PlatformUninstallResp)
    - [PlatformUpgradeReq](#cc.arduino.cli.commands.PlatformUpgradeReq)
    - [PlatformUpgradeResp](#cc.arduino.cli.commands.PlatformUpgradeResp)
  
  
  
  

- [commands/lib.proto](#commands/lib.proto)
    - [DownloadResource](#cc.arduino.cli.commands.DownloadResource)
    - [InstalledLibrary](#cc.arduino.cli.commands.InstalledLibrary)
    - [Library](#cc.arduino.cli.commands.Library)
    - [Library.PropertiesEntry](#cc.arduino.cli.commands.Library.PropertiesEntry)
    - [LibraryDependency](#cc.arduino.cli.commands.LibraryDependency)
    - [LibraryDependencyStatus](#cc.arduino.cli.commands.LibraryDependencyStatus)
    - [LibraryDownloadReq](#cc.arduino.cli.commands.LibraryDownloadReq)
    - [LibraryDownloadResp](#cc.arduino.cli.commands.LibraryDownloadResp)
    - [LibraryInstallReq](#cc.arduino.cli.commands.LibraryInstallReq)
    - [LibraryInstallResp](#cc.arduino.cli.commands.LibraryInstallResp)
    - [LibraryListReq](#cc.arduino.cli.commands.LibraryListReq)
    - [LibraryListResp](#cc.arduino.cli.commands.LibraryListResp)
    - [LibraryRelease](#cc.arduino.cli.commands.LibraryRelease)
    - [LibraryResolveDependenciesReq](#cc.arduino.cli.commands.LibraryResolveDependenciesReq)
    - [LibraryResolveDependenciesResp](#cc.arduino.cli.commands.LibraryResolveDependenciesResp)
    - [LibrarySearchReq](#cc.arduino.cli.commands.LibrarySearchReq)
    - [LibrarySearchResp](#cc.arduino.cli.commands.LibrarySearchResp)
    - [LibraryUninstallReq](#cc.arduino.cli.commands.LibraryUninstallReq)
    - [LibraryUninstallResp](#cc.arduino.cli.commands.LibraryUninstallResp)
    - [LibraryUpgradeAllReq](#cc.arduino.cli.commands.LibraryUpgradeAllReq)
    - [LibraryUpgradeAllResp](#cc.arduino.cli.commands.LibraryUpgradeAllResp)
    - [SearchedLibrary](#cc.arduino.cli.commands.SearchedLibrary)
    - [SearchedLibrary.ReleasesEntry](#cc.arduino.cli.commands.SearchedLibrary.ReleasesEntry)
  
    - [LibraryLayout](#cc.arduino.cli.commands.LibraryLayout)
    - [LibraryLocation](#cc.arduino.cli.commands.LibraryLocation)
  
  
  

- [commands/upload.proto](#commands/upload.proto)
    - [UploadReq](#cc.arduino.cli.commands.UploadReq)
    - [UploadResp](#cc.arduino.cli.commands.UploadResp)
  
  
  
  

- [Scalar Value Types](#scalar-value-types)



<a name="commands/board.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## commands/board.proto



<a name="cc.arduino.cli.commands.BoardAttachReq"></a>

### BoardAttachReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.commands.Instance) |  |  |
| board_uri | [string](#string) |  |  |
| sketch_path | [string](#string) |  |  |
| search_timeout | [string](#string) |  |  |






<a name="cc.arduino.cli.commands.BoardAttachResp"></a>

### BoardAttachResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| task_progress | [TaskProgress](#cc.arduino.cli.commands.TaskProgress) |  |  |






<a name="cc.arduino.cli.commands.BoardDetailsReq"></a>

### BoardDetailsReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.commands.Instance) |  |  |
| fqbn | [string](#string) |  |  |






<a name="cc.arduino.cli.commands.BoardDetailsResp"></a>

### BoardDetailsResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| config_options | [ConfigOption](#cc.arduino.cli.commands.ConfigOption) | repeated |  |
| required_tools | [RequiredTool](#cc.arduino.cli.commands.RequiredTool) | repeated |  |






<a name="cc.arduino.cli.commands.BoardListAllReq"></a>

### BoardListAllReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.commands.Instance) |  |  |
| search_args | [string](#string) | repeated |  |






<a name="cc.arduino.cli.commands.BoardListAllResp"></a>

### BoardListAllResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| boards | [BoardListItem](#cc.arduino.cli.commands.BoardListItem) | repeated |  |






<a name="cc.arduino.cli.commands.BoardListItem"></a>

### BoardListItem



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| FQBN | [string](#string) |  |  |






<a name="cc.arduino.cli.commands.BoardListReq"></a>

### BoardListReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.commands.Instance) |  |  |






<a name="cc.arduino.cli.commands.BoardListResp"></a>

### BoardListResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ports | [DetectedPort](#cc.arduino.cli.commands.DetectedPort) | repeated |  |






<a name="cc.arduino.cli.commands.ConfigOption"></a>

### ConfigOption



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| option | [string](#string) |  |  |
| option_label | [string](#string) |  |  |
| values | [ConfigValue](#cc.arduino.cli.commands.ConfigValue) | repeated |  |






<a name="cc.arduino.cli.commands.ConfigValue"></a>

### ConfigValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [string](#string) |  |  |
| value_label | [string](#string) |  |  |
| selected | [bool](#bool) |  |  |






<a name="cc.arduino.cli.commands.DetectedPort"></a>

### DetectedPort



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| address | [string](#string) |  |  |
| protocol | [string](#string) |  |  |
| protocol_label | [string](#string) |  |  |
| boards | [BoardListItem](#cc.arduino.cli.commands.BoardListItem) | repeated |  |






<a name="cc.arduino.cli.commands.RequiredTool"></a>

### RequiredTool



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| version | [string](#string) |  |  |
| packager | [string](#string) |  |  |





 

 

 

 



<a name="commands/commands.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## commands/commands.proto



<a name="cc.arduino.cli.commands.DestroyReq"></a>

### DestroyReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.commands.Instance) |  |  |






<a name="cc.arduino.cli.commands.DestroyResp"></a>

### DestroyResp







<a name="cc.arduino.cli.commands.InitReq"></a>

### InitReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| library_manager_only | [bool](#bool) |  |  |






<a name="cc.arduino.cli.commands.InitResp"></a>

### InitResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.commands.Instance) |  |  |
| platforms_index_errors | [string](#string) | repeated |  |
| libraries_index_error | [string](#string) |  |  |
| download_progress | [DownloadProgress](#cc.arduino.cli.commands.DownloadProgress) |  |  |
| task_progress | [TaskProgress](#cc.arduino.cli.commands.TaskProgress) |  |  |






<a name="cc.arduino.cli.commands.RescanReq"></a>

### RescanReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.commands.Instance) |  |  |






<a name="cc.arduino.cli.commands.RescanResp"></a>

### RescanResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| platforms_index_errors | [string](#string) | repeated |  |
| libraries_index_error | [string](#string) |  |  |






<a name="cc.arduino.cli.commands.UpdateIndexReq"></a>

### UpdateIndexReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.commands.Instance) |  |  |






<a name="cc.arduino.cli.commands.UpdateIndexResp"></a>

### UpdateIndexResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| download_progress | [DownloadProgress](#cc.arduino.cli.commands.DownloadProgress) |  |  |






<a name="cc.arduino.cli.commands.UpdateLibrariesIndexReq"></a>

### UpdateLibrariesIndexReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.commands.Instance) |  |  |






<a name="cc.arduino.cli.commands.UpdateLibrariesIndexResp"></a>

### UpdateLibrariesIndexResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| download_progress | [DownloadProgress](#cc.arduino.cli.commands.DownloadProgress) |  |  |






<a name="cc.arduino.cli.commands.VersionReq"></a>

### VersionReq







<a name="cc.arduino.cli.commands.VersionResp"></a>

### VersionResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| version | [string](#string) |  |  |





 

 

 


<a name="cc.arduino.cli.commands.ArduinoCore"></a>

### ArduinoCore
The main Arduino Platform Service

BOOTSTRAP COMMANDS
-------------------

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Init | [InitReq](#cc.arduino.cli.commands.InitReq) | [InitResp](#cc.arduino.cli.commands.InitResp) stream | Start a new instance of the Arduino Core Service |
| Destroy | [DestroyReq](#cc.arduino.cli.commands.DestroyReq) | [DestroyResp](#cc.arduino.cli.commands.DestroyResp) | Destroy an instance of the Arduino Core Service |
| Rescan | [RescanReq](#cc.arduino.cli.commands.RescanReq) | [RescanResp](#cc.arduino.cli.commands.RescanResp) | Rescan instance of the Arduino Core Service |
| UpdateIndex | [UpdateIndexReq](#cc.arduino.cli.commands.UpdateIndexReq) | [UpdateIndexResp](#cc.arduino.cli.commands.UpdateIndexResp) stream | Update package index of the Arduino Core Service |
| UpdateLibrariesIndex | [UpdateLibrariesIndexReq](#cc.arduino.cli.commands.UpdateLibrariesIndexReq) | [UpdateLibrariesIndexResp](#cc.arduino.cli.commands.UpdateLibrariesIndexResp) stream | Update libraries index |
| Version | [VersionReq](#cc.arduino.cli.commands.VersionReq) | [VersionResp](#cc.arduino.cli.commands.VersionResp) |  |
| BoardDetails | [BoardDetailsReq](#cc.arduino.cli.commands.BoardDetailsReq) | [BoardDetailsResp](#cc.arduino.cli.commands.BoardDetailsResp) | Requests details about a board |
| BoardAttach | [BoardAttachReq](#cc.arduino.cli.commands.BoardAttachReq) | [BoardAttachResp](#cc.arduino.cli.commands.BoardAttachResp) stream |  |
| BoardList | [BoardListReq](#cc.arduino.cli.commands.BoardListReq) | [BoardListResp](#cc.arduino.cli.commands.BoardListResp) |  |
| BoardListAll | [BoardListAllReq](#cc.arduino.cli.commands.BoardListAllReq) | [BoardListAllResp](#cc.arduino.cli.commands.BoardListAllResp) |  |
| Compile | [CompileReq](#cc.arduino.cli.commands.CompileReq) | [CompileResp](#cc.arduino.cli.commands.CompileResp) stream |  |
| PlatformInstall | [PlatformInstallReq](#cc.arduino.cli.commands.PlatformInstallReq) | [PlatformInstallResp](#cc.arduino.cli.commands.PlatformInstallResp) stream |  |
| PlatformDownload | [PlatformDownloadReq](#cc.arduino.cli.commands.PlatformDownloadReq) | [PlatformDownloadResp](#cc.arduino.cli.commands.PlatformDownloadResp) stream |  |
| PlatformUninstall | [PlatformUninstallReq](#cc.arduino.cli.commands.PlatformUninstallReq) | [PlatformUninstallResp](#cc.arduino.cli.commands.PlatformUninstallResp) stream |  |
| PlatformUpgrade | [PlatformUpgradeReq](#cc.arduino.cli.commands.PlatformUpgradeReq) | [PlatformUpgradeResp](#cc.arduino.cli.commands.PlatformUpgradeResp) stream |  |
| Upload | [UploadReq](#cc.arduino.cli.commands.UploadReq) | [UploadResp](#cc.arduino.cli.commands.UploadResp) stream |  |
| PlatformSearch | [PlatformSearchReq](#cc.arduino.cli.commands.PlatformSearchReq) | [PlatformSearchResp](#cc.arduino.cli.commands.PlatformSearchResp) |  |
| PlatformList | [PlatformListReq](#cc.arduino.cli.commands.PlatformListReq) | [PlatformListResp](#cc.arduino.cli.commands.PlatformListResp) |  |
| LibraryDownload | [LibraryDownloadReq](#cc.arduino.cli.commands.LibraryDownloadReq) | [LibraryDownloadResp](#cc.arduino.cli.commands.LibraryDownloadResp) stream |  |
| LibraryInstall | [LibraryInstallReq](#cc.arduino.cli.commands.LibraryInstallReq) | [LibraryInstallResp](#cc.arduino.cli.commands.LibraryInstallResp) stream |  |
| LibraryUninstall | [LibraryUninstallReq](#cc.arduino.cli.commands.LibraryUninstallReq) | [LibraryUninstallResp](#cc.arduino.cli.commands.LibraryUninstallResp) stream |  |
| LibraryUpgradeAll | [LibraryUpgradeAllReq](#cc.arduino.cli.commands.LibraryUpgradeAllReq) | [LibraryUpgradeAllResp](#cc.arduino.cli.commands.LibraryUpgradeAllResp) stream |  |
| LibraryResolveDependencies | [LibraryResolveDependenciesReq](#cc.arduino.cli.commands.LibraryResolveDependenciesReq) | [LibraryResolveDependenciesResp](#cc.arduino.cli.commands.LibraryResolveDependenciesResp) |  |
| LibrarySearch | [LibrarySearchReq](#cc.arduino.cli.commands.LibrarySearchReq) | [LibrarySearchResp](#cc.arduino.cli.commands.LibrarySearchResp) |  |
| LibraryList | [LibraryListReq](#cc.arduino.cli.commands.LibraryListReq) | [LibraryListResp](#cc.arduino.cli.commands.LibraryListResp) |  |

 



<a name="commands/common.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## commands/common.proto



<a name="cc.arduino.cli.commands.DownloadProgress"></a>

### DownloadProgress



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| url | [string](#string) |  |  |
| file | [string](#string) |  |  |
| total_size | [int64](#int64) |  |  |
| downloaded | [int64](#int64) |  |  |
| completed | [bool](#bool) |  |  |






<a name="cc.arduino.cli.commands.Instance"></a>

### Instance



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int32](#int32) |  |  |






<a name="cc.arduino.cli.commands.TaskProgress"></a>

### TaskProgress



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| message | [string](#string) |  |  |
| completed | [bool](#bool) |  |  |





 

 

 

 



<a name="commands/compile.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## commands/compile.proto



<a name="cc.arduino.cli.commands.CompileReq"></a>

### CompileReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.commands.Instance) |  |  |
| fqbn | [string](#string) |  | Fully Qualified Board Name, e.g.: arduino:avr:uno. |
| sketchPath | [string](#string) |  |  |
| showProperties | [bool](#bool) |  | Show all build preferences used instead of compiling. |
| preprocess | [bool](#bool) |  | Print preprocessed code to stdout. |
| buildCachePath | [string](#string) |  | Builds of &#39;core.a&#39; are saved into this path to be cached and reused. |
| buildPath | [string](#string) |  | Path where to save compiled files. |
| buildProperties | [string](#string) | repeated | List of custom build properties separated by commas. Or can be used multiple times for multiple properties. |
| warnings | [string](#string) |  | Used to tell gcc which warning level to use. |
| verbose | [bool](#bool) |  | Turns on verbose mode. |
| quiet | [bool](#bool) |  | Suppresses almost every output. |
| vidPid | [string](#string) |  | VID/PID specific build properties. |
| exportFile | [string](#string) |  | The compiled binary is written to this file |
| jobs | [int32](#int32) |  | The max number of concurrent compiler instances to run (as make -jx) |
| libraries | [string](#string) | repeated | List of custom libraries paths separated by commas. Or can be used multiple times for multiple libraries paths. |
| optimizeForDebug | [bool](#bool) |  | Optimize compile output for debug, not for release |






<a name="cc.arduino.cli.commands.CompileResp"></a>

### CompileResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| out_stream | [bytes](#bytes) |  |  |
| err_stream | [bytes](#bytes) |  |  |





 

 

 

 



<a name="commands/core.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## commands/core.proto



<a name="cc.arduino.cli.commands.Board"></a>

### Board



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| fqbn | [string](#string) |  |  |






<a name="cc.arduino.cli.commands.Platform"></a>

### Platform



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ID | [string](#string) |  |  |
| Installed | [string](#string) |  |  |
| Latest | [string](#string) |  |  |
| Name | [string](#string) |  |  |
| Maintainer | [string](#string) |  |  |
| Website | [string](#string) |  |  |
| Email | [string](#string) |  |  |
| Boards | [Board](#cc.arduino.cli.commands.Board) | repeated |  |






<a name="cc.arduino.cli.commands.PlatformDownloadReq"></a>

### PlatformDownloadReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.commands.Instance) |  |  |
| platform_package | [string](#string) |  |  |
| architecture | [string](#string) |  |  |
| version | [string](#string) |  |  |






<a name="cc.arduino.cli.commands.PlatformDownloadResp"></a>

### PlatformDownloadResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| progress | [DownloadProgress](#cc.arduino.cli.commands.DownloadProgress) |  |  |






<a name="cc.arduino.cli.commands.PlatformInstallReq"></a>

### PlatformInstallReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.commands.Instance) |  |  |
| platform_package | [string](#string) |  |  |
| architecture | [string](#string) |  |  |
| version | [string](#string) |  |  |






<a name="cc.arduino.cli.commands.PlatformInstallResp"></a>

### PlatformInstallResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| progress | [DownloadProgress](#cc.arduino.cli.commands.DownloadProgress) |  |  |
| task_progress | [TaskProgress](#cc.arduino.cli.commands.TaskProgress) |  |  |






<a name="cc.arduino.cli.commands.PlatformListReq"></a>

### PlatformListReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.commands.Instance) |  |  |
| updatable_only | [bool](#bool) |  |  |






<a name="cc.arduino.cli.commands.PlatformListResp"></a>

### PlatformListResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| installed_platform | [Platform](#cc.arduino.cli.commands.Platform) | repeated |  |






<a name="cc.arduino.cli.commands.PlatformSearchReq"></a>

### PlatformSearchReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.commands.Instance) |  |  |
| search_args | [string](#string) |  |  |
| all_versions | [bool](#bool) |  |  |






<a name="cc.arduino.cli.commands.PlatformSearchResp"></a>

### PlatformSearchResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| search_output | [Platform](#cc.arduino.cli.commands.Platform) | repeated |  |






<a name="cc.arduino.cli.commands.PlatformUninstallReq"></a>

### PlatformUninstallReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.commands.Instance) |  |  |
| platform_package | [string](#string) |  |  |
| architecture | [string](#string) |  |  |






<a name="cc.arduino.cli.commands.PlatformUninstallResp"></a>

### PlatformUninstallResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| task_progress | [TaskProgress](#cc.arduino.cli.commands.TaskProgress) |  |  |






<a name="cc.arduino.cli.commands.PlatformUpgradeReq"></a>

### PlatformUpgradeReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.commands.Instance) |  |  |
| platform_package | [string](#string) |  |  |
| architecture | [string](#string) |  |  |






<a name="cc.arduino.cli.commands.PlatformUpgradeResp"></a>

### PlatformUpgradeResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| progress | [DownloadProgress](#cc.arduino.cli.commands.DownloadProgress) |  |  |
| task_progress | [TaskProgress](#cc.arduino.cli.commands.TaskProgress) |  |  |





 

 

 

 



<a name="commands/lib.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## commands/lib.proto



<a name="cc.arduino.cli.commands.DownloadResource"></a>

### DownloadResource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| url | [string](#string) |  |  |
| archivefilename | [string](#string) |  |  |
| checksum | [string](#string) |  |  |
| size | [int64](#int64) |  |  |
| cachepath | [string](#string) |  |  |






<a name="cc.arduino.cli.commands.InstalledLibrary"></a>

### InstalledLibrary



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| library | [Library](#cc.arduino.cli.commands.Library) |  |  |
| release | [LibraryRelease](#cc.arduino.cli.commands.LibraryRelease) |  |  |






<a name="cc.arduino.cli.commands.Library"></a>

### Library



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| author | [string](#string) |  |  |
| maintainer | [string](#string) |  |  |
| sentence | [string](#string) |  |  |
| paragraph | [string](#string) |  |  |
| website | [string](#string) |  |  |
| category | [string](#string) |  |  |
| architectures | [string](#string) | repeated |  |
| types | [string](#string) | repeated |  |
| install_dir | [string](#string) |  |  |
| source_dir | [string](#string) |  |  |
| utility_dir | [string](#string) |  |  |
| location | [string](#string) |  |  |
| container_platform | [string](#string) |  |  |
| layout | [string](#string) |  |  |
| real_name | [string](#string) |  |  |
| dot_a_linkage | [bool](#bool) |  |  |
| precompiled | [bool](#bool) |  |  |
| ld_flags | [string](#string) |  |  |
| is_legacy | [bool](#bool) |  |  |
| version | [string](#string) |  |  |
| license | [string](#string) |  |  |
| properties | [Library.PropertiesEntry](#cc.arduino.cli.commands.Library.PropertiesEntry) | repeated |  |






<a name="cc.arduino.cli.commands.Library.PropertiesEntry"></a>

### Library.PropertiesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="cc.arduino.cli.commands.LibraryDependency"></a>

### LibraryDependency



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| version_constraint | [string](#string) |  |  |






<a name="cc.arduino.cli.commands.LibraryDependencyStatus"></a>

### LibraryDependencyStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| versionRequired | [string](#string) |  |  |
| versionInstalled | [string](#string) |  |  |






<a name="cc.arduino.cli.commands.LibraryDownloadReq"></a>

### LibraryDownloadReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.commands.Instance) |  |  |
| name | [string](#string) |  |  |
| version | [string](#string) |  |  |






<a name="cc.arduino.cli.commands.LibraryDownloadResp"></a>

### LibraryDownloadResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| progress | [DownloadProgress](#cc.arduino.cli.commands.DownloadProgress) |  |  |






<a name="cc.arduino.cli.commands.LibraryInstallReq"></a>

### LibraryInstallReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.commands.Instance) |  |  |
| name | [string](#string) |  |  |
| version | [string](#string) |  |  |






<a name="cc.arduino.cli.commands.LibraryInstallResp"></a>

### LibraryInstallResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| progress | [DownloadProgress](#cc.arduino.cli.commands.DownloadProgress) |  |  |
| task_progress | [TaskProgress](#cc.arduino.cli.commands.TaskProgress) |  |  |






<a name="cc.arduino.cli.commands.LibraryListReq"></a>

### LibraryListReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.commands.Instance) |  |  |
| all | [bool](#bool) |  |  |
| updatable | [bool](#bool) |  |  |






<a name="cc.arduino.cli.commands.LibraryListResp"></a>

### LibraryListResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| installed_library | [InstalledLibrary](#cc.arduino.cli.commands.InstalledLibrary) | repeated |  |






<a name="cc.arduino.cli.commands.LibraryRelease"></a>

### LibraryRelease



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| author | [string](#string) |  |  |
| version | [string](#string) |  |  |
| maintainer | [string](#string) |  |  |
| sentence | [string](#string) |  |  |
| paragraph | [string](#string) |  |  |
| website | [string](#string) |  |  |
| category | [string](#string) |  |  |
| architectures | [string](#string) | repeated |  |
| types | [string](#string) | repeated |  |
| resources | [DownloadResource](#cc.arduino.cli.commands.DownloadResource) |  |  |
| license | [string](#string) |  |  |
| provides_includes | [string](#string) | repeated |  |
| dependencies | [LibraryDependency](#cc.arduino.cli.commands.LibraryDependency) | repeated |  |






<a name="cc.arduino.cli.commands.LibraryResolveDependenciesReq"></a>

### LibraryResolveDependenciesReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.commands.Instance) |  |  |
| name | [string](#string) |  |  |
| version | [string](#string) |  |  |






<a name="cc.arduino.cli.commands.LibraryResolveDependenciesResp"></a>

### LibraryResolveDependenciesResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| dependencies | [LibraryDependencyStatus](#cc.arduino.cli.commands.LibraryDependencyStatus) | repeated |  |






<a name="cc.arduino.cli.commands.LibrarySearchReq"></a>

### LibrarySearchReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.commands.Instance) |  |  |
| query | [string](#string) |  |  |






<a name="cc.arduino.cli.commands.LibrarySearchResp"></a>

### LibrarySearchResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| libraries | [SearchedLibrary](#cc.arduino.cli.commands.SearchedLibrary) | repeated |  |






<a name="cc.arduino.cli.commands.LibraryUninstallReq"></a>

### LibraryUninstallReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.commands.Instance) |  |  |
| name | [string](#string) |  |  |
| version | [string](#string) |  |  |






<a name="cc.arduino.cli.commands.LibraryUninstallResp"></a>

### LibraryUninstallResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| task_progress | [TaskProgress](#cc.arduino.cli.commands.TaskProgress) |  |  |






<a name="cc.arduino.cli.commands.LibraryUpgradeAllReq"></a>

### LibraryUpgradeAllReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.commands.Instance) |  |  |






<a name="cc.arduino.cli.commands.LibraryUpgradeAllResp"></a>

### LibraryUpgradeAllResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| progress | [DownloadProgress](#cc.arduino.cli.commands.DownloadProgress) |  |  |
| task_progress | [TaskProgress](#cc.arduino.cli.commands.TaskProgress) |  |  |






<a name="cc.arduino.cli.commands.SearchedLibrary"></a>

### SearchedLibrary



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| releases | [SearchedLibrary.ReleasesEntry](#cc.arduino.cli.commands.SearchedLibrary.ReleasesEntry) | repeated |  |
| latest | [LibraryRelease](#cc.arduino.cli.commands.LibraryRelease) |  |  |






<a name="cc.arduino.cli.commands.SearchedLibrary.ReleasesEntry"></a>

### SearchedLibrary.ReleasesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [LibraryRelease](#cc.arduino.cli.commands.LibraryRelease) |  |  |





 


<a name="cc.arduino.cli.commands.LibraryLayout"></a>

### LibraryLayout


| Name | Number | Description |
| ---- | ------ | ----------- |
| flat_layout | 0 |  |
| recursive_layout | 1 |  |



<a name="cc.arduino.cli.commands.LibraryLocation"></a>

### LibraryLocation


| Name | Number | Description |
| ---- | ------ | ----------- |
| ide_builtin | 0 |  |
| platform_builtin | 1 |  |
| referenced_platform_builtin | 2 |  |
| sketchbook | 3 |  |


 

 

 



<a name="commands/upload.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## commands/upload.proto



<a name="cc.arduino.cli.commands.UploadReq"></a>

### UploadReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [Instance](#cc.arduino.cli.commands.Instance) |  |  |
| fqbn | [string](#string) |  |  |
| sketch_path | [string](#string) |  |  |
| port | [string](#string) |  |  |
| verbose | [bool](#bool) |  |  |
| verify | [bool](#bool) |  |  |
| import_file | [string](#string) |  |  |






<a name="cc.arduino.cli.commands.UploadResp"></a>

### UploadResp



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| out_stream | [bytes](#bytes) |  |  |
| err_stream | [bytes](#bytes) |  |  |





 

 

 

 



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

