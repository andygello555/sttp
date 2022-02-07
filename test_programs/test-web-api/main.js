const cluster = require('cluster')
const http = require('http');

const hostname = '127.0.0.1';
const port = 3000;
const workers = 12;

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
        switch (resObj.query_params.format) {
            case 'html':
                res.setHeader('Content-Type', 'text/html')
                let query_params = '<li>query_params:<ul>'
                for (const [param, value] of Object.entries(resObj.query_params)) {
                    query_params += `<li>${param}: ${value}</li>`
                }
                query_params += `</ul></li>`
                let headers = '<li>headers:<ul>'
                for (const [name, value] of Object.entries(resObj.headers)) {
                    headers += `<li>${name}: ${value}</li>`
                }
                headers += `</ul></li>`
                let html = `<html lang="en">
    <head><title>${resObj.method}: ${resObj.url}</title></head>
    <body>
        <h1>${resObj.method}: ${resObj.url}</h1>
        <div>
            <ul>
                <li>method: ${resObj.method}</li>
                <li>url: ${resObj.url}</li>
                ${query_params}
                ${headers}
                <li>code: ${resObj.code}</li>
                <li>version: ${resObj.version}</li>`

                if (resObj.body) {
                    html += `<li>body:<div>${JSON.stringify(resObj.body)}</div></li>`
                }
                html += '</ul></div></body></html>'
                res.end(html)
                break
            case 'json':
            default:
                res.setHeader('Content-Type', 'application/json');
                res.end(JSON.stringify(resObj))
                break
        }
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
    console.log(`worker ${process.pid} started`)
}
