import {
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  TableCaption,
  Button,
} from "@chakra-ui/react";
import { PhoneIcon, AddIcon, WarningIcon, MinusIcon } from "@chakra-ui/icons";

import { useReducer, useEffect } from "react";

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

const ACTION_TYPES = {
  INCREMENT_ITEM: "incItem",
  DECREMENT_ITEM: "decItem",
  ADD_ITEM: "addItem",
};

const reducer = (state, action) => {
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
  }
};

const CTable = (itemArray) => {
  var items =
    Object.keys(itemArray).length === 0 ? sampleItemsArray : itemArray;

  const [state, dispatch] = useReducer(reducer, items);
  var itemIds = Object.keys(state);
  var itemAttributes = Object.keys(state[itemIds[0]]);
  const itemAdd = (id) => {
    dispatch({ type: ACTION_TYPES.INCREMENT_ITEM, payload: id });
  };
  const itemSubtract = (id) => {
    dispatch({ type: ACTION_TYPES.DECREMENT_ITEM, payload: id });
  };
  return (
    <Table variant="simple">
      <TableCaption>Shopping Cart</TableCaption>
      <Thead>
        <Tr>
          {itemAttributes.map((n) => (
            <Th>{n}</Th>
          ))}
          <Th align="centre">Change</Th>
        </Tr>
      </Thead>
      <Tbody>
        {itemIds.map((n) => (
          <Tr>
            {itemAttributes.map((x) => (
              <Td>{state[n][x]}</Td>
            ))}
            <Td>
              <Button marginRight={5} onClick={(e) => itemAdd(n)}>
                <AddIcon />
              </Button>
              <Button onClick={(e) => itemSubtract(n)}>
                <MinusIcon />
              </Button>
            </Td>
          </Tr>
        ))}
      </Tbody>
    </Table>
  );
};

export default CTable;
