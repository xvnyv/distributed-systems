export const mergeVersions = (badgerObject) => {
  if (badgerObject.Versions.length <= 1) {
    return badgerObject.Versions[0];
  }
  var result = badgerObject.Versions[0];
  for (let i = 1; i < badgerObject.Versions.length; i++) {
    result = mergeClientCarts(result, badgerObject.Versions[i]);
  }
  return result;
};

const mergeClientCarts = (clientCartSelf, clientCartReceived) => {
  var output = {};
  output.UserID = clientCartReceived.UserID;
  var newmap = {};
  var clientCartItemIds = Object.keys(clientCartSelf.Item);
  for (var i = 0; i < clientCartItemIds.length; i++) {
    var currentKey = clientCartItemIds[i];
    var currentObject = clientCartSelf.Item[currentKey];
    if (clientCartReceived.Item[currentKey] !== undefined) {
      if (
        currentObject.Quantity < clientCartReceived.Item[currentKey].Quantity
      ) {
        currentObject.Quantity = clientCartReceived.Item[currentKey].Quantity;
      }
    }
    newmap[currentKey] = currentObject;
  }

  var clientCartReceivedItemIds = Object.keys(clientCartReceived.Item);
  for (var i = 0; i < clientCartReceivedItemIds.length; i++) {
    var currentKey = clientCartReceivedItemIds[i];
    if (!newmap[currentKey]) {
      newmap[currentKey] = clientCartReceived.Item[currentKey];
    }
  }

  output.Item = newmap;

  var newVectorClock = {};
  var selfVectorClockKeys = Object.keys(clientCartSelf.VectorClock);
  for (var i = 0; i < selfVectorClockKeys.length; i++) {
    Max(clientCartSelf.VectorClock[i], clientCartReceived.VectorClock[i]);
  }

  for (var key in clientCartSelf.VectorClock) {
    if (clientCartReceived.VectorClock[key] === undefined) {
      newVectorClock[key] = clientCartSelf.VectorClock[key];
      continue;
    }
    newVectorClock[key] = Max(
      clientCartSelf.VectorClock[key],
      clientCartReceived.VectorClock[key]
    );
  }

  for (var key in clientCartReceived.VectorClock) {
    if (newVectorClock[key] === undefined) {
      newVectorClock[key] = clientCartReceived.VectorClock[key];
    }
  }

  output.VectorClock = newVectorClock;

  return output;
};

const Max = (v1, v2) => {
  return v1 < v2 ? v2 : v1;
};
