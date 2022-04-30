import { SendPostRequest } from "../http_helpers/PostGetRequesters";

export const clientCartReducer = (state, action) => {
  switch (action.type) {
    case CLIENTCART_ACTIONS.UPDATE_VECTORCLOCK:
      var newState = {
        ...state,
        VectorClock: action.payload,
      };
      return newState;
    case CLIENTCART_ACTIONS.UPDATE_STATE:
      return action.payload;
  }

  return newState;
};

export const CLIENTCART_ACTIONS = {
  UPDATE_VECTORCLOCK: "updateVectorClock",
  UPDATE_STATE: "updateState",
};
