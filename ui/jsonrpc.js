// Taken from https://github.com/corps/angular-golang-json-rpc-client
// License is unknown so far
function JsonRpcClientProvider($httpProvider) {
	var requestCounter = 0;
	var serviceUrl = "";
	var services = {};

	this.addService = function(serviceName, methods) {
		if (services[serviceName]) {
			throw "Json Rpc Service " + serviceName + " has already been registered!";
		}
		services[serviceName] = methods;
	}

	this.setServiceEndpoint = function(url) {
		serviceUrl = url;
	}

	this.$get = ['$q', '$http', function($q, $http){
		var client = {};

		var successHandler = function(result) {
			if(result.data.error) {
				return $q.reject(result.data.error);
			}
			return result.data.result;
		}

		var failureHandler = function(result) {
			return $q.reject("Got HTTP Error " + result.status + ": " + result);
		}

		function serviceMethodFactory(serviceName, methodName) {
			return function(requestObj) {
				var payload = {};
				payload.method = serviceName + "." + methodName;
				payload.id = ++requestCounter;
				payload.params = [requestObj];

				var config = {};
				config.headers = {};
				config.headers['Content-Type'] = "application/json";
				config.headers['Accept'] = "application/json";

				return $http.post(serviceUrl, JSON.stringify(payload), config).then(successHandler, failureHandler);
			};
		}

		for(var serviceName in services) {
			var service = {};
			var methods = services[serviceName];
			for(var i = 0; i < methods.length; ++i) {
				service[methods[i]] = serviceMethodFactory(serviceName, methods[i]);
			}
			client[serviceName] = service;
		}
		return client;
	}];
}

angular.module('corps.jsonrpc', [])
	.provider('jsonRpcClient', ['$httpProvider', JsonRpcClientProvider]);
