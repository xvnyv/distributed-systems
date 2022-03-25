export const reducer = (state, action) => {
  switch (action.type) {
    case ACTION_TYPES.INCREMENT_ITEM:
      var id = action.payload;
      return {
        ...state,
        [id]: { ...state[id], Quantity: state[id].Quantity + 1 },
      };
    case ACTION_TYPES.DECREMENT_ITEM:
      var id = action.payload;
      if (state[id].Quantity === 0) return state;
      return {
        ...state,
        [id]: { ...state[id], Quantity: state[id].Quantity - 1 },
      };
    case ACTION_TYPES.ADD_ITEM:
      return {
        ...state,
        [action.payload.id]: { ...state[id], Quantity: state[id].Quantity - 1 },
      };
  }
};

export const ACTION_TYPES = {
  INCREMENT_ITEM: "incItem",
  DECREMENT_ITEM: "decItem",
  ADD_ITEM: "addItem",
};


