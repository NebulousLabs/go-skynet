# Skynet Go SDK

An SDK for integrating Skynet into Go applications.

## Examples

### Upload a file

```go
package main

import (
	"fmt"
	skynet "github.com/NebulousLabs/go-skynet"
)

func main() {
	skylink, err := skynet.UploadFile("./image.jpg", skynet.DefaultUploadOptions)
	if err != nil {
		fmt.Println("Unable to upload:", err.Error())
		return
	}
	fmt.Printf("Upload successful, skylink: %v\n", skylink)
}
```

### Download a file

```go
package main

import (
	"fmt"
	skynet "github.com/NebulousLabs/go-skynet"
)

func main() {
	// Must have a 'skylink' from an earlier upload.

	err = skynet.DownloadFile("./dst.go", skylink, skynet.DefaultDownloadOptions)
	if err != nil {
		fmt.Println("Something went wrong, please try again.\nError:", err.Error())
		return
	}
	fmt.Println("Download successful")
}
```

### Upload a directory

```go
package main

import (
	"fmt"
	skynet "github.com/NebulousLabs/go-skynet"
)

func main() {
	url, err := skynet.UploadDirectory("./images", skynet.DefaultUploadOptions)
	if err != nil {
		fmt.Println("Unable to upload:", err.Error())
		return
	}
	fmt.Printf("Upload successful, url: %v\n", url)
}
```
