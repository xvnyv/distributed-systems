import CAddItemModal from "./components/CAddItemModal";
import CTable from "./components/CTable";
import {
  Container,
  Grid,
  GridItem,
  Box,
  Input,
  FormLabel,
  Button,
  useToast,
} from "@chakra-ui/react";
import React, { useReducer, useState } from "react";
import { ITEM_ACTIONS, reducer } from "./reducers/ItemReducer";

var clientCartObject = {
  userId: 5,
  item: {
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
  },
  vectorClock: [1, 2, 3, 4],
};

function App() {
  const [state, dispatch] = useReducer(reducer, clientCartObject);

  const toast = useToast();
  const toastIdRef = React.useRef();

  const CheckEnter = (e) => {
    console.log(e.target.value);
    if (e.key === "Enter") {
      console.log("hi");
      findUser(e.target.value);
      return;
    }
    if (isNaN(e.key)) {
      toastIdRef.current = toast({
        title: "Input Error",
        status: "warning",
        description: "User's Id should be a number",
        duration: 2000,
        isClosable: true,
        position: "top-left",
      });
    }
  };

  const findUser = (userId) => {
    console.log(userId);
    dispatch({ type: ITEM_ACTIONS.CHANGE_USER, payload: userId });
  };

  return (
    <div>
      <Grid templateColumns="repeat(6, 1fr)" gap={2}>
        <GridItem colSpan={2} padding={3}>
          <Box borderRadius={10} border="1px" borderColor="blackAlpha.100" padding={5}>
            <CAddItemModal state={state} dispatch={dispatch} />
            <FormLabel marginTop={10} marginLeft={2}>
              Change User
            </FormLabel>
            <Input placeholder="e.g, '5'" onKeyPress={(e) => CheckEnter(e)} type="number" />
            <Button marginTop={3} width="100%">
              Find User
            </Button>
          </Box>
        </GridItem>
        <GridItem colSpan={4} padding={3}>
          <Box borderRadius={10} border="1px" borderColor="blackAlpha.100" padding={5}>
            {/* Table for to show shopping cart items */}
            <CTable state={state} dispatch={dispatch} />
          </Box>
        </GridItem>
      </Grid>
    </div>
  );
}

export default App;
