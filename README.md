# cftool

A lightweight tool that makes working with CloudFormation easier. You can learn
more about CloudFormation on the [AWS website](https://aws.amazon.com/cloudformation/).

## Installation

If you have a valid Golang setup (GOPATH, etc.), you can simply run:

```
go install github.com/ElasticProjects/cftool
```

and cftool will install to your GOPATH.

It's our plan to provide an installation method for non-golang users as well.

## Why YAML?

The core of cftool is the idea that using YAML with some helpful macros can
significantly ease development of CloudFormation scripts. YAML is helpful for
a few reasons:

1. It's easier to write YAML than JSON, since YAML enforces good indentation
and requires a lot fewer keystrokes.
1. YAML includes the idea of tags, which we use in cftool to add extra functionality
to templates.
1. YAML is ultimately a superset of JSON, so it's pretty easy to convert back
to JSON.

## Getting Started

A quick caveat - cftool is intended for use by people who already understand
how to use CloudFormation. Be sure you understand the core concepts behind
CloudFormation before you attempt to use cftool.

To get started, let's initialize a new cftool project with `cftool init demo`.
This will create the `demo` folder if it doesn't already exist, and create some
files under it - `imports/`, `files/`, and `config.yml`.

Now we'll set up a really simple template that starts a new EC2 instance.
Create a file named `demo.yml` with the following data:

```yaml
---
Parameters:
  InstanceType:
    Description: The type of instance you'd like to start
    Type: String
    Default: t2.nano

Resources:
  Instance:
    Properties:
      ImageId: ami-08111162
      InstanceType: !ref InstanceType
      KeyName: default
```

Let's run through what we're doing above. This simple CloudFormation template
defines a parameter, `InstanceType`, and a resource, the EC2 instances we're
going to start. As you can see in the InstanceType property under the
resource, we're referencing the `InstanceType` parameter. That's your first
cftool tag - `!ref [Reference]` is a shorthand form for `{ "Reference" : "[Reference]" }`
that you frequently use in CloudFormation Templates.

Now that we've got a template, let's process it to see the JSON it generates:

```
cftool process demo.yml > demo.json
```

This command will process the template, converting it to JSON, and write it
to a file named `demo.json`. If you open the file you'll see this:

```json
{
  "Parameters": {
    "InstanceType": {
      "Default": "t2.nano",
      "Description": "The type of instance you'd like to start",
      "Type": "String"
    }
  },
  "Resources": {
    "Instance": {
      "Properties": {
        "ImageId": "ami-08111162",
        "InstanceType": {
          "Ref": "InstanceType"
        },
        "KeyName": "default"
      }
    }
  }
}
```

Congratulations, you've created your first CloudFormation template using cftool!

## Tags

cftool includes several helpful tags. Here's a list of them and examples
of using them.

!import

!ref

!file

!vault

!config

