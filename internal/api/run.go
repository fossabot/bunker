package api

import (
    "os"
    "fmt"
    "github.com/containerd/containerd"
    "github.com/containerd/containerd/oci"
    "github.com/containerd/containerd/cio"

    lib "github.com/coditva/bunker/internal"
    util "github.com/coditva/bunker/internal/util"
    types "github.com/coditva/bunker/internal/types"
)

func (api Api) Run(args *types.Args, reply *string) error {
    imageName := (*args)[0]
    runCommand := (*args)[1]

    if imageName == "" {
        *reply = "No image to run from"
        lib.Logger.Warning(*reply)
        return nil
    }

    if runCommand == "" {
        *reply = "No command given to run"
        lib.Logger.Warning(*reply)
        return nil
    }

    image, err := lib.ContainerdClient.Client.Pull(lib.ContainerdClient.Ns, imageName, containerd.WithPullUnpack)
    if err != nil {
        lib.Logger.Error(err)
        return err
    }

    id := util.NewRandomName()
    container, err := lib.ContainerdClient.Client.NewContainer(lib.ContainerdClient.Ns, id,
            containerd.WithNewSnapshot(id, image),
            containerd.WithNewSpec(oci.WithImageConfig(image)))
    if err != nil {
        lib.Logger.Error(err)
        return nil
    }

    os.Mkdir(lib.ContainerStreamBasePath, 0660)

    containerOut, err := os.OpenFile(fmt.Sprintf("%v%v.out", lib.ContainerStreamBasePath, id),
            os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
    if err != nil {
        lib.Logger.Error(err)
        return err
    }

    containerIn, err := os.OpenFile(fmt.Sprintf("%v%v.in", lib.ContainerStreamBasePath, id),
            os.O_CREATE|os.O_RDONLY|os.O_APPEND, 0660)
    if err != nil {
        lib.Logger.Error(err)
        return err
    }

    containerErr, err := os.OpenFile(fmt.Sprintf("%v%v.err", lib.ContainerStreamBasePath, id),
            os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
    if err != nil {
        lib.Logger.Error(err)
        return err
    }

    task, err := container.NewTask(lib.ContainerdClient.Ns,
            cio.NewCreator(cio.WithStreams(containerIn, containerOut, containerErr)))
    if err != nil {
        *reply = "Could not create new task"
        lib.Logger.Error(err)
    }
    defer task.Delete(lib.ContainerdClient.Ns)
    task.Start(lib.ContainerdClient.Ns)

    lib.Logger.Info(fmt.Sprintf("Running command %v on image %v in container ID %v",
            runCommand, imageName, container.ID()))

    *reply = fmt.Sprintf("%v", id)

    return nil
}
