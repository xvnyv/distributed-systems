export const mergeVersions = (badgerObject) => {
  if (badgerObject.Versions.length <= 1) {
    return badgerObject.Versions[0];
  }
  // get the latest of each client
  //{ clientId: clientCart}
  var versionsToMerge = {};

  for (var i = 0; i < badgerObject.Versions.length; i++) {
    var version = badgerObject.Versions[i];
    console.log("version in badgerobject.versions", version);
    if (versionsToMerge[version.ClientId] !== undefined) {
      //check whether version should override
      const d1 = versionsToMerge[version.ClientId].Timestamp;
      const d2 = version.Timestamp;
      console.log("d1<d2: ", d1 < d2);
      if (d1 < d2) {
        //override
        versionsToMerge[version.ClientId] = version;
      }
      continue;
    }
    //get max vectorclock
    version.VectorClock = mergeVectorClock(
      versionsToMerge[version.ClientId],
      version
    );
    //add
    versionsToMerge[version.ClientId] = version;
  }

  var firstKey = Object.keys(versionsToMerge)[0];
  var result = versionsToMerge[firstKey];
  console.log("versionstomerge,", versionsToMerge);
  var keylst = Object.keys(versionsToMerge);
  for (var i = 1; i < keylst.length; i++) {
    console.log(versionsToMerge[keylst[i]]);
    result = mergeClientCarts(result, versionsToMerge[keylst[i]]);
  }
  return result;
};

const mergeClientCarts = (clientCartSelf, clientCartReceived) => {
  var output = {};
  console.log("clientcartreceived:", clientCartReceived);
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

  var newVectorClock = mergeVectorClock(clientCartSelf, clientCartReceived);

  output.VectorClock = newVectorClock;
  return output;
};

const Max = (v1, v2) => {
  return v1 < v2 ? v2 : v1;
};

const mergeVectorClock = (clientCart1, clientCart2) => {
  var newVectorClock = {};
  var selfVectorClockKeys = Object.keys(clientCart1.VectorClock);
  for (var i = 0; i < selfVectorClockKeys.length; i++) {
    Max(clientCart1.VectorClock[i], clientCart2.VectorClock[i]);
  }

  for (var key in clientCart1.VectorClock) {
    if (clientCart2.VectorClock[key] === undefined) {
      newVectorClock[key] = clientCart1.VectorClock[key];
      continue;
    }
    newVectorClock[key] = Max(
      clientCart1.VectorClock[key],
      clientCart2.VectorClock[key]
    );
  }

  return newVectorClock;
};
