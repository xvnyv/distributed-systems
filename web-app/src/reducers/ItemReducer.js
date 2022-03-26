export const reducer = (state, action) => {
  switch (action.type) {
    case ITEM_ACTIONS.INCREMENT:
      var id = action.payload;
      return {
        ...state,
        item: {
          ...state.item,
          [id]: { ...state.item[id], Quantity: state.item[id].Quantity++ },
        },
      };
    case ITEM_ACTIONS.DECREMENT:
      var id = action.payload;
      return {
        ...state,
        item: {
          ...state.item,
          [id]: { ...state.item[id], Quantity: state.item[id].Quantity-- },
        },
      };
    case ITEM_ACTIONS.NEW:
      return {
        ...state,
        item: {
          ...state.item,
          [action.payload.itemId]: {
            id: action.payload.itemId,
            Name: action.payload.itemName,
            Quantity: 1,
          },
        },
      };
    case ITEM_ACTIONS.DELETE:
      var id = action.payload;
      delete state.item[id];
      return state;
    case ITEM_ACTIONS.CHANGE_USER:
      // findUser(action.payload);
      return state;
  }
};

export const ITEM_ACTIONS = {
  INCREMENT: "incItem",
  DECREMENT: "decItem",
  NEW: "addItem",
  DELETE: "delItem",
  CHANGE_USER: "changeUser",
};
