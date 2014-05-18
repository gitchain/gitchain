angular.module('gitchain', ['corps.jsonrpc']).
  config(['jsonRpcClientProvider', function(clientProvider) {
  	clientProvider.setServiceEndpoint("/rpc")
  	clientProvider.addService('KeyService', ['GeneratePrivateKey', 'ListPrivateKeys', 'SetMainKey', 'GetMainKey'])
  }]).
  controller('MainController', ['$scope', '$http', '$timeout', 'jsonRpcClient',
  function($scope, $http, $timeout,api) {

    $scope.privateKeys = []
    $scope.mainPrivateKey = null

    var loadPrivateKeys = function() {
      api.KeyService.ListPrivateKeys({}).then(function(data) {
        $scope.privateKeys = (data.Aliases || [])
        $scope.privateKeys.push(["Create a new one..."])
      })
      api.KeyService.GetMainKey({}).then(function(data) {
        $scope.mainPrivateKey = data.Alias
      })
    }
    loadPrivateKeys()

    $scope.$watch('mainPrivateKey', function() {
      if ($scope.mainPrivateKey == "Create a new one...") {
        bootbox.prompt("New private key alias", function(alias) {
          if (alias !== null) {
            api.KeyService.GeneratePrivateKey({Alias: alias}).then(function(data) {
              loadPrivateKeys()
              bootbox.dialog({
                message: "Private key pair named " + alias + " has been generated",
                title: "Alias created",
                buttons: {
                  confirms: {
                    label: "OK",
                    className: "btn-success",
                  }
                }
              })
            })
          } else {
          }
        })
      } else {
        api.KeyService.SetMainKey({alias: $scope.mainPrivateKey})
      }
    })

    $scope.info = {}
    var requestInfo = function() {
      $http({method: 'GET', url: '/info'}).success(function(data) {
         $scope.info = data
         $timeout(requestInfo, 500)
      })
    }
    requestInfo()

  }]);
