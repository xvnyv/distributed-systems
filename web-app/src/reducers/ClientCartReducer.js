import { SendPostRequest } from "../http_helpers/PostGetRequesters";

export const clientCartReducer = (state, action) => {
  switch (action.type) {
    case ITEM_ACTIONS.INCREMENT:
      var id = action.payload.id;
      var newState = {
        ...state,
        Item: {
          ...state.Item,
          [id]: { ...state.Item[id], Quantity: state.Item[id].Quantity + 1 },
        },
      };

      break;
    case ITEM_ACTIONS.DECREMENT:
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
    case ITEM_ACTIONS.NEW:
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
    case ITEM_ACTIONS.DELETE:
      var id = action.payload.toDelete;
      delete state.Item[id];
      var newState = state;
      break;
    case ITEM_ACTIONS.CHANGE_USER:
      return action.payload;
  }

  const result = SendPostRequest(
    newState,
    action.payload.toast,
    action.payload.toastRef,
    action.payload.dispatch
  );

  return newState;
};

export const ITEM_ACTIONS = {
  INCREMENT: "incItem",
  DECREMENT: "decItem",
  NEW: "addItem",
  DELETE: "delItem",
  CHANGE_USER: "changeUser",
  NEW_USER: "newUser",
};
