import axios from "axios";
import { v4 as uuidv4 } from "uuid";
import { ITEM_ACTIONS } from "../reducers/ItemReducer";

const GetUserData = async (userId, dispatch) => {
  const res = await fetch(`http://localhost:8080/read-request?id=${userId}`, {
    method: "GET",
    headers: {
      "Content-type": "application/json",
    },
  })
    .then((response) => {
      response.json().then((data) => {
        console.log("Success:", data.Data);
        dispatch({ type: ITEM_ACTIONS.CHANGE_USER, payload: data.Data });
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

export default GetUserData;

export const SendPostRequest = async (item) => {
  item["Guid"] = uuidv4();
  console.log(item);
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
      console.log("Success:", data);
    })
    .catch((error) => {
      console.error("Error:", error);
    });
};
