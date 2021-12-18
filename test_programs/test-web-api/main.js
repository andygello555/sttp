const cluster = require('cluster')
const http = require('http');

const hostname = '127.0.0.1';
const port = 3000;
const workers = 6;

// https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods
const REQUEST_HAS_BODY_METHODS = [
    'POST',
    'PUT',
    'DELETE',
    'PATCH'
]

const paramsToObject = (searchParams) => {
    const result = {}
    for (const [key, value] of searchParams.entries()) {
        result[key] = value
    }
    return result
}

const responder = (req, res) => {
    console.log(`replying with worker ${process.pid}`)
    // We always 200 as the server is just a simple echo server
    res.statusCode = 200;
    res.setHeader('Content-Type', 'application/json');

    // Defining our response object that will be stringified
    const fullUrl = `http://${hostname}:${port}${req.url}`
    let resObj = {
        'method': req.method,
        'url': fullUrl,
        'query_params': paramsToObject(new URL(fullUrl).searchParams),
        'headers': req.headers,
        'code': req.statusCode,
        'version': req.httpVersion,
    }

    const send = () => {
        res.end(JSON.stringify(resObj))
    }

    // We add in the request body if it's one of the HTTP methods that should have a body
    if (REQUEST_HAS_BODY_METHODS.includes(req.method)) {
        let body = "";

        // We read in the request body to the body variable
        req.on('readable', () => {
            let next = req.read()
            if (next) {
                body += next
            }
        })

        // When we have read in all of the request body we will encode it to the correct content-type, add it to the
        // response object and send the response object
        req.on('end', () => {
            let encoded = body
            if (req.headers['content-type']) {
                switch (req.headers["content-type"].toLowerCase()) {
                    case 'application/json':
                        encoded = JSON.parse(body)
                        break
                    default:
                        encoded = body
                        break
                }
            }
            resObj['body'] = encoded
            send()
        })
    } else {
        send()
    }
}

if (cluster.isMaster) {
    console.log(`primary ${process.pid} is running`)

    for (let i = 0; i < workers; i++) {
        cluster.fork()
    }

    cluster.on('exit', (worker, code, signal) => {
        console.log(`worker ${worker.process.pid} died`)
    })
} else {
    http.createServer(responder).listen(port)
    //     , hostname, () => {
    //     console.log(`Server running at http://${hostname}:${port}/`);
    // })
    // server.listen(port, hostname, () => {
    // });
    console.log(`worker ${process.pid} started`)
}
