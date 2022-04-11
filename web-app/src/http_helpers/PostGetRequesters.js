//Helper functions
import { CLIENTCART_ACTIONS } from "../reducers/ClientCartReducer";
import { mergeVersions } from "../utils/merge";

export const SendGetRequest = async (userId, dispatch, toast, toastIdRef) => {
  await fetch(`http://localhost:8080/read-request?id=${userId}`, {
    method: "GET",
    headers: {
      "Content-type": "application/json",
    },
  })
    .then((response) => {
      console.log(response);
      if (response.status >= 502) {
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
          if (data.Status === 0) {
            toastIdRef.current = toast({
              title: "Error",
              status: "error",
              description: "Get Request Error : " + data.Error,
              duration: 2000,
              isClosable: true,
              position: "top-right",
            });
          }
          if (data.Data.UserID === "") {
            dispatch({
              type: CLIENTCART_ACTIONS.UPDATE_STATE,
              payload: { UserID: userId, Item: {} },
            });
          } else {
            dispatch({
              type: CLIENTCART_ACTIONS.UPDATE_STATE,
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
            type: CLIENTCART_ACTIONS.UPDATE_STATE,
            payload: { UserID: userId, Item: {} },
          });
        });
    })
    .catch((err) => {
      console.log("ERROR: ", err.toString());
      toastIdRef.current = toast({
        title: "Error",
        status: "error",
        description: err.toString(),
        duration: 2000,
        isClosable: true,
        position: "top-right",
      });
    });
  // const data = await res.json();
};

export const SendPostRequest = async (
  item,
  toast,
  toastIdRef,
  dispatch,
  clientId
) => {
  item.ClientId = clientId;
  item.Timestamp = Date.now();
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
        type: CLIENTCART_ACTIONS.UPDATE_VECTORCLOCK,
        payload: data.Data.Versions[0].VectorClock,
      });
    })
    .catch((error) => {
      console.error("Error:", error);
    });
};
