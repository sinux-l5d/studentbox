## Compile

First, make sure you have go 1.20 installed, along with `make`.

Then, download dependencies:
```
make download
```

and build the project:
```
make build
```

## Use

You must have a working installation of Podman, 4.0.0 or later.

you can then interact with the CLI:
```
./bin/studentbox spawn -u <username> -p <projectname> -r <runtimename>
```

Where `<runtimename>` is the name of a directory in the `runtimes` directory.

## AWS

If you want to try this on AWS, two files are provided to help you get started:
- `aws/vpc.yaml` is a CloudFormation template that will create a VPC and a EC2 instance, preconfigured with Podman and the studentbox CLI. Just make shure to run the command prompted on first connection
- `aws/launch.sh` is a utility script that will create a CloudFormation stack and wait for it's creation. It also format the parameters in a simpler way. To deploy, make sure you have a working AWS CLI installation, and run:
    ```
    ./aws/launch.sh <stackname> KeyName=<sshkeyname> 
    ```
