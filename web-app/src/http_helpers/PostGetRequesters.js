// dependencies
import { v4 as uuidv4 } from "uuid";
//Helper functions
import { ITEM_ACTIONS } from "../reducers/ClientCartReducer";

export const SendGetRequest = async (userId, dispatch) => {
  const res = await fetch(`http://localhost:8080/read-request?id=${userId}`, {
    method: "GET",
    headers: {
      "Content-type": "application/json",
    },
  })
    .then((response) => {
      response.json().then((data) => {
        console.log("Success:", data.Data);
        if (data.Data.UserID === "") {
          dispatch({
            type: ITEM_ACTIONS.CHANGE_USER,
            payload: { UserID: userId, Item: {} },
          });
        } else {
          dispatch({ type: ITEM_ACTIONS.CHANGE_USER, payload: data.Data });
        }
      });
    })
    .catch((err) => {
      console.log(err);
      dispatch({
        type: ITEM_ACTIONS.CHANGE_USER,
        payload: { UserID: userId, Item: {} },
      });
    });
  // const data = await res.json();
};

export const SendPostRequest = async (item, toast, toastIdRef, dispatch) => {
  item["Guid"] = uuidv4();
  console.log(item);
  console.log("sending");
  fetch(`http://localhost:8080/write-request`, {
    method: "POST",
    headers: {
      "Content-type": "application/json",
    },
    body: JSON.stringify(item),
  })
    .then((response) => response.json())
    .then((data) => {
      console.log("success: ", data);
      toastIdRef.current = toast({
        title: "Saved",
        status: "success",
        description:
          "User's cart updated with vector clock :" + data.Data.VectorClock,
        duration: 2000,
        isClosable: true,
        position: "bottom-left",
      });
      dispatch({ type: ITEM_ACTIONS.CHANGE_USER, payload: data.Data });
    })
    .catch((error) => {
      console.error("Error:", error);
    });
};
