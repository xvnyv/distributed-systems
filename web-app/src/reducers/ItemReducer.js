export const reducer = (state, action) => {
  switch (action.type) {
    case ITEM_ACTIONS.INCREMENT:
      var id = action.payload;
      return {
        ...state,
        [id]: { ...state[id], Quantity: state[id].Quantity++ },
      };
    case ITEM_ACTIONS.DECREMENT:
      var id = action.payload;
      return {
        ...state,
        [id]: { ...state[id], Quantity: state[id].Quantity-- },
      };
    case ITEM_ACTIONS.NEW:
      return {
        ...state,
        [action.payload.itemId]: {
          id: action.payload.itemId,
          Name: action.payload.itemName,
          Quantity: 1,
        },
      };
    case ITEM_ACTIONS.DELETE:
      var id = action.payload;
      delete state[id];
      return state;
  }
};

export const ITEM_ACTIONS = {
  INCREMENT: "incItem",
  DECREMENT: "decItem",
  NEW: "addItem",
  DELETE: "delItem",
};
