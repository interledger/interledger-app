export default {
	async fetch(request, env, context) {
		const url = new URL(request.url);
		const key = url.pathname.slice(1);

		if (!authorizeRequest(request, env)) {
			return new Response('Forbidden', { status: 403 });
		}

		switch (request.method) {
			case 'PUT':
				await env.BUCKET.put(key, request.body, {
					httpMetadata: request.headers,
				});
				return new Response(`Put ${key} successfully!`);
			case 'GET':
				return handleGet(url, request, env.BUCKET, context)
			case 'OPTIONS':
				return handleOptions(request)
			default:
				return new Response('Method Not Allowed', {
					status: 405,
					headers: {
						Allow: 'PUT, GET, DELETE',
					},
				});
		}
	},
};

async function handleGet(url, request, bucket, context) {
	// Construct the cache key from the cache URL
	const cacheKey = new Request(url.toString(), request);
	const cache = await caches.open('cdn');

	// Check whether the value is already available in the cache
	// if not, you will need to fetch it from R2, and store it in the cache
	// for future access
	let response = await cache.match(cacheKey);

	if (response) {
		return response;
	}

	const key = url.pathname.slice(1);
	const object = await bucket.get(key);
	if (object === null) {
		return objectNotFound(key);
	}

	const headers = new Headers();
	object.writeHttpMetadata(headers);
	headers.set('etag', object.httpEtag);

	// Cache API respects Cache-Control headers. Setting s-max-age to 10
	// will limit the response to be in cache for 10 seconds max
	// Any changes made to the response here will be reflected in the cached value
	headers.append('Cache-Control', 's-maxage=31536000');

	response = new Response(object.body, {
		headers,
	});

	context.waitUntil(cache.put(cacheKey.clone(), response.clone()));

	return response
}

function objectNotFound(objectName) {
	return new Response(`<html><body>Fynbos CDN Object "<b>${objectName}</b>" not found</body></html>`, {
		status: 404,
		headers: {
			'content-type': 'text/html; charset=UTF-8'
		}
	})
}

const hasValidHeader = (request, env) => {
	return request.headers.get('X-Fynbos-Auth-Key') === env.AUTH_KEY_SECRET;
};

function authorizeRequest(request, env) {
	switch (request.method) {
		case 'PUT':
		case 'DELETE':
			return hasValidHeader(request, env);
		case 'GET':
			return true;
		case 'OPTIONS':
			return true;
		default:
			return false;
	}
}

// We support the GET, POST, HEAD, and OPTIONS methods from any origin,
// and accept the Content-Type header on requests. These headers must be
// present on all responses to all CORS requests. In practice, this means
// all responses to OPTIONS or POST requests.
const corsHeaders = {
	"Access-Control-Allow-Origin": "*",
	"Access-Control-Allow-Methods": "GET, HEAD, POST, OPTIONS",
	"Access-Control-Allow-Headers": "Content-Type",
}

function handleOptions(request) {
	if (request.headers.get("Origin") !== null &&
		request.headers.get("Access-Control-Request-Method") !== null &&
		request.headers.get("Access-Control-Request-Headers") !== null) {
		// Handle CORS pre-flight request.
		return new Response(null, {
			headers: corsHeaders
		})
	} else {
		// Handle standard OPTIONS request.
		return new Response(null, {
			headers: {
				"Allow": "GET, POST, OPTIONS",
			}
		})
	}
}