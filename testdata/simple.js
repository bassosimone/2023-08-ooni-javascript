const time = require("golang/time")
const dsl = require("ooni/dsl")

const pipeline = dsl.compose(
    dsl.domainName("www.youtube.com"),
    dsl.dnsLookupGetaddrinfo(),
    dsl.makeEndpointsForPort(443),
    dsl.newEndpointPipeline(
        dsl.compose(
            dsl.tcpConnect(),
            dsl.tlsHandshake(),
            dsl.httpConnectionTls(),
            dsl.httpTransaction(),
            dsl.discard(),
        )
    )
)

const zeroTime = time.now()
console.log("current time:", zeroTime.Format("2006-01-02T15:04:05.999999999Z07:00"))

const result = dsl.run(pipeline, zeroTime)
console.log(JSON.stringify(result))
