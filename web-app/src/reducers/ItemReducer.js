import GetDataAxios from "../axios/GetData";

export const reducer = (state, action) => {
  switch (action.type) {
    case ITEM_ACTIONS.INCREMENT:
      var id = action.payload;
      console.log({
        ...state,
        Item: {
          ...state.Item,
          [id]: { ...state.Item[id], Quantity: state.Item[id].Quantity + 1 },
        },
      });
      return {
        ...state,
        Item: {
          ...state.Item,
          [id]: { ...state.Item[id], Quantity: state.Item[id].Quantity + 1 },
        },
      };
    case ITEM_ACTIONS.DECREMENT:
      var id = action.payload;
      return {
        ...state,
        Item: {
          ...state.Item,
          [id]: { ...state.Item[id], Quantity: state.Item[id].Quantity - 1 },
        },
      };
    case ITEM_ACTIONS.NEW:
      return {
        ...state,
        Item: {
          ...state.Item,
          [action.payload.itemId]: {
            id: action.payload.itemId,
            Name: action.payload.itemName,
            Quantity: 1,
          },
        },
      };
    case ITEM_ACTIONS.DELETE:
      var id = action.payload;
      delete state.Item[id];
      return state;
    case ITEM_ACTIONS.CHANGE_USER:
      return action.payload;
  }
};

export const ITEM_ACTIONS = {
  INCREMENT: "incItem",
  DECREMENT: "decItem",
  NEW: "addItem",
  DELETE: "delItem",
  CHANGE_USER: "changeUser",
};
