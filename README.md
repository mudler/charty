# Charty - charts based runner framework

Example: 

```bash
charty run --set bar=fff --set foo=aa --run 'commands[0].run=bash test.sh' --run 'commands[0].name=clitest' test/fixture
charty run --set bar=fff --set foo=aa --run 'commands[0].run=bash test.sh' --run 'commands[0].name=clitest' https://...tgz
charty run --set bar=fff --set foo=aa --run 'commands[0].run=bash test.sh' --run 'commands[0].name=clitest' tests.tgz

```