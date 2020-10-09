# Charty - Runnable helm-style charts runner

Charty is a dead-simple and no-frills chart runner expecially well-suited to package and release testsuites next to your software.

Why? 
- Create portable and versioned charts that can be released and tracked next to your software
- Template in helm-style to allow override of values and runtime options when executing test
- Simple - does one thing and does it good, it is just a template/runner mechanism, nothing else
- Easy to embed - can be your entrypoint for docker images that run your test in a reproducible environment
- Syntax/usage inspired from helm, merge values and override easily from CLI

Have you wondered to package and version set of blackbox commands in a templated-way, in a helm fashion?

Example: 

```bash
charty run --set bar=fff --set foo=aa --run 'commands[0].run=bash test.sh' --run 'commands[0].name=clitest' test/fixture
charty run --set bar=fff --set foo=aa --run 'commands[0].run=bash test.sh' --run 'commands[0].name=clitest' https://...tgz
charty run --set bar=fff --set foo=aa --run 'commands[0].run=bash test.sh' --run 'commands[0].name=clitest' tests.tgz

```

How a chart looks like?

```bash
chart/
    templates/ #All files under this directory gets templated
        file.sh
        foo.yaml
        bar.js
        baz.rb
    static/ # Files that are copied as-is in the execution runtime
        notemplated.sh
    metadata.yaml # Chart metadata
    runtime.yaml # Runtime options that can be override from cli
    values.yaml # Default values used for template interpolation
```

## Run charts

Running a chart is as easy as executing `charty run`. It takes only one argument and it's the chart path (local directory, URLs, and `tar.gz` compressed archives are supported). The chart values can be override with ```--values-files``` and runtime options can be override with ```--run-files```. To note, each single value in the yamls can be override by cli, with ```--set key=value``` and ```--run key=value```

## Package charts

Charty can be used to package a chart, although it's a merely compression of a chart folder.

You can run 
```
charty package <localchart> <destination_dir>
```

To generate a new chart.

### Generate templated charts for debugging

You can run 
```
charty template <localchart> <destination_dir>
```

to write a generated version of the local chart for local inspection