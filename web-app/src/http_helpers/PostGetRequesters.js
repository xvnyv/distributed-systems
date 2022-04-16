// dependencies
import { v4 as uuidv4 } from "uuid";
//Helper functions
import { CLIENTCART_ACTIONS } from "../reducers/ClientCartReducer";
import { mergeVersions } from "../utils/merge";

export const SendGetRequest = async (userId, dispatch, toast, toastIdRef) => {
  console.log("error: ");
  await fetch(`http://localhost:8080/read-request?id=${userId}`, {
    method: "GET",
    headers: {
      "Content-type": "application/json",
    },
  })
    .then((response) => {
      if (response.status >= 500) {
        toastIdRef.current = toast({
          title: "Error",
          status: "error",
          description: "Get Request Error :" + response.statusText,
          duration: 2000,
          isClosable: true,
          position: "top-right",
        });
        return;
      }
      response
        .json()
        .then((data) => {
          console.log("Success:", data);
          if (data.Data.UserID === "") {
            dispatch({
              type: CLIENTCART_ACTIONS.CHANGE_USER,
              payload: { UserID: userId, Item: {} },
            });
          } else {
            dispatch({
              type: CLIENTCART_ACTIONS.CHANGE_USER,
              payload: mergeVersions(data.Data),
            });
          }
        })
        .catch((err) => {
          console.log("ERROR: ", err);
          toastIdRef.current = toast({
            title: "Error",
            status: "error",
            description: "Get Request Error :" + err,
            duration: 2000,
            isClosable: true,
            position: "top-right",
          });
          dispatch({
            type: CLIENTCART_ACTIONS.CHANGE_USER,
            payload: { UserID: userId, Item: {} },
          });
        });
    })
    .catch((err) => {
      console.log("ERROR: ", err);
      dispatch({
        type: CLIENTCART_ACTIONS.CHANGE_USER,
        payload: { UserID: userId, Item: {} },
      });
    });
  // const data = await res.json();
};

export const SendPostRequest = async (item, toast, toastIdRef, dispatch) => {
  item["Guid"] = uuidv4();
  console.log("sending");
  await fetch(`http://localhost:8080/write-request`, {
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
          "User's cart updated with vector clock :" +
          JSON.stringify(data.Data.Versions[0].VectorClock),
        duration: 2000,
        isClosable: true,
        position: "bottom-left",
      });
      dispatch({
        type: CLIENTCART_ACTIONS.UPDATE_STATE,
        payload: data.Data.Versions[0],
      });
    })
    .catch((error) => {
      console.error("Error:", error);
    });
};