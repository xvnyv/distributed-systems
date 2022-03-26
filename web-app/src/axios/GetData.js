import axios from "axios";
import { ITEM_ACTIONS } from "../reducers/ItemReducer";

const GetDataAxios = async (userId, dispatch) => {
  axios.get(`http://localhost:8080/read?id=${userId}`).then((value) => {
    dispatch({ type: ITEM_ACTIONS.CHANGE_USER, payload: value.data.Data });
  });
};

export default GetDataAxios;
