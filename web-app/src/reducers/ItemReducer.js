import axios from "axios";
import GetDataAxios, { SendPostRequest } from "../axios/GetData";

export const reducer = (state, action) => {
  switch (action.type) {
    case ITEM_ACTIONS.INCREMENT:
      var id = action.payload;
      var newState = {
        ...state,
        Item: {
          ...state.Item,
          [id]: { ...state.Item[id], Quantity: state.Item[id].Quantity + 1 },
        },
      };

      break;
    case ITEM_ACTIONS.DECREMENT:
      var id = action.payload;
      var newState = {
        ...state,
        Item: {
          ...state.Item,
          [id]: { ...state.Item[id], Quantity: state.Item[id].Quantity - 1 },
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
      var id = action.payload;
      delete state.Item[id];
      var newState = state;
      break;
    case ITEM_ACTIONS.CHANGE_USER:
      var newState = action.payload;
      break;
  }
  if (action.type !== ITEM_ACTIONS.CHANGE_USER) {
    SendPostRequest(newState);
  }
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
