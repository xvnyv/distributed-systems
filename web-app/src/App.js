import CModal from "./components/CModal";
import CTable from "./components/CTable";
import logo from "./logo.svg";

import { useReducer } from "react";
import { reducer } from "./reducers/ItemReducer";

var sampleItemsArray = {
  1: {
    id: 1,
    Name: "pencil",
    Quantity: 2,
  },
  3: {
    id: 3,
    Name: "paper",
    Quantity: 1,
  },
};

function App(itemArray) {
  var items =
    Object.keys(itemArray).length === 0 ? sampleItemsArray : itemArray;
  const [state, dispatch] = useReducer(reducer, items);

  return (
    <div>
      <CModal state={state} dispatch={dispatch}/>
      <CTable state={state} dispatch={dispatch} />
    </div>
  );
}

export default App;
