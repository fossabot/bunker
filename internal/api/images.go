package api

import (
    "fmt"

    lib "github.com/coditva/bunker/internal"
    types "github.com/coditva/bunker/internal/types"
)

func (api Api) Images(args *types.Args, reply *string) error {
    lib.Logger.Info("Getting images from containerd")
    images, err := lib.ContainerdClient.Client.ListImages(lib.ContainerdClient.Ns, "")
    if err != nil {
        return err
    }

    for _, image := range images{
        name := image.Name()
        size, err := image.Size(lib.ContainerdClient.Ns)
        if err != nil {
            lib.Logger.Warning("Unknown size for image ", name, ": ", err)
            size = 0
        }
        *reply = fmt.Sprintf("%v%v\t%v\n", *reply, name, size)
    }

    return nil
}