angular.module('gitchain', ['corps.jsonrpc']).
  config(['jsonRpcClientProvider', function(clientProvider) {
  	clientProvider.setServiceEndpoint("/rpc")
  	clientProvider.addService('KeyService', ['GeneratePrivateKey', 'ListPrivateKeys', 'SetMainKey', 'GetMainKey'])
    clientProvider.addService('BlockService', ['GetBlock','BlockTransactions'])
  }])
  .directive('gcBlock', function() {
    return {
      restrict: 'E'
    };
  })
  .controller('MainController', ['$scope', '$http', '$timeout', 'jsonRpcClient',
  function($scope, $http, $timeout,api) {

    $scope.block = function(block) {
      console.log(block);
    }
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
              $.notify("Private key `" + alias + "' has been generated", "success")
            })
          } else {
          }
        })
      } else {
        api.KeyService.SetMainKey({alias: $scope.mainPrivateKey})
      }
    })

    $scope.blocks = []
    $scope.info = {}
    var requestInfo = function() {
      $http({method: 'GET', url: '/info'}).success(function(data) {
         $scope.info = data
         $timeout(requestInfo, 500)
         if ($scope.blocks.length == 0 || $scope.blocks[0].Hash != $scope.info.LastBlock) {
           var hash = $scope.info.LastBlock
           var newBlocks = []
           var handleBlock = function(data) {
              data["Hash"] = hash
              newBlocks.push(data)
              hash = data.PreviousBlockHash
              if (hash != "0000000000000000000000000000000000000000000000000000000000000000" && newBlocks.length < 5) {
                api.BlockService.GetBlock({Hash: hash}).then(handleBlock)
              } else {
                $scope.blocks = newBlocks
              }
           }
           api.BlockService.GetBlock({Hash: hash}).then(handleBlock)
         }
      })
    }
    requestInfo()

  }]);
