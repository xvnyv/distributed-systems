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

import {reducer, ACTION_TYPES} from "../reducers/ItemReducer";








const CTable = ({state, dispatch}) => {

  console.log(state)
  console.log(dispatch)
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
