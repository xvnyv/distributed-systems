import { SendPostRequest } from "../http_helpers/PostGetRequesters";

export const clientCartReducer = (state, action) => {
  switch (action.type) {
    case CLIENTCART_ACTIONS.INCREMENT:
      var id = action.payload.id;
      var newState = {
        ...state,
        Item: {
          ...state.Item,
          [id]: { ...state.Item[id], Quantity: state.Item[id].Quantity + 1 },
        },
      };

      break;
    case CLIENTCART_ACTIONS.DECREMENT:
      var id = action.payload.id;
      var newQty = state.Item[id].Quantity - 1;
      var newState = {
        ...state,
        Item: {
          ...state.Item,
          [id]: { ...state.Item[id], Quantity: newQty },
        },
      };
      break;
    case CLIENTCART_ACTIONS.NEW:
      var newState = {
        ...state,
        Item: {
          ...state.Item,
          [parseInt(action.payload.itemId)]: {
            Id: parseInt(action.payload.itemId),
            Name: action.payload.itemName,
            Quantity: 1,
          },
        },
      };
      break;
    case CLIENTCART_ACTIONS.DELETE:
      var id = action.payload.toDelete;
      delete state.Item[id];
      var newState = state;
      break;
    case CLIENTCART_ACTIONS.CHANGE_USER:
      return action.payload;
    case CLIENTCART_ACTIONS.UPDATE_STATE:
      if (vectorClockSmaller(state.VectorClock, action.payload.VectorClock)) {
        console.log("overwrite possible");
        return action.payload;
      }
      return state;
  }

  const result = SendPostRequest(
    newState,
    action.payload.toast,
    action.payload.toastRef,
    action.payload.dispatch
  );

  return newState;
};

export const CLIENTCART_ACTIONS = {
  INCREMENT: "incItem",
  DECREMENT: "decItem",
  NEW: "addItem",
  DELETE: "delItem",
  CHANGE_USER: "changeUser",
  UPDATE_STATE: "updateState",
};

const vectorClockSmaller = (arr1, arr2) => {
  for (var i = 0; i < arr2.length; i++) {
    if (arr1[i] > arr2[i]) {
      return false;
    }
  }
  return true;
};
