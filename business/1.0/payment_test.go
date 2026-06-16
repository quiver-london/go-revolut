package business

import (
	"encoding/json"
	"testing"
)

func TestPaymentReceiverOmitsEmptyAccountId(t *testing.T) {
	body, err := json.Marshal(PaymentReq{
		RequestId: "request-id",
		AccountId: "source-account-id",
		Receiver: PaymentReceiver{
			CounterpartyId: "counterparty-id",
		},
		Amount:   12.34,
		Currency: "GBP",
	})
	if err != nil {
		t.Fatalf("marshal payment request: %v", err)
	}

	var payment map[string]interface{}
	if err := json.Unmarshal(body, &payment); err != nil {
		t.Fatalf("unmarshal payment request: %v", err)
	}

	receiver, ok := payment["receiver"].(map[string]interface{})
	if !ok {
		t.Fatalf("receiver has unexpected type %T", payment["receiver"])
	}

	if _, ok := receiver["account_id"]; ok {
		t.Fatalf("receiver.account_id should be omitted when empty: %s", body)
	}
}

func TestPaymentReceiverIncludesAccountId(t *testing.T) {
	body, err := json.Marshal(PaymentReq{
		RequestId: "request-id",
		AccountId: "source-account-id",
		Receiver: PaymentReceiver{
			CounterpartyId: "counterparty-id",
			AccountId:      "receiver-account-id",
		},
		Amount:   12.34,
		Currency: "GBP",
	})
	if err != nil {
		t.Fatalf("marshal payment request: %v", err)
	}

	var payment map[string]interface{}
	if err := json.Unmarshal(body, &payment); err != nil {
		t.Fatalf("unmarshal payment request: %v", err)
	}

	receiver, ok := payment["receiver"].(map[string]interface{})
	if !ok {
		t.Fatalf("receiver has unexpected type %T", payment["receiver"])
	}

	if got := receiver["account_id"]; got != "receiver-account-id" {
		t.Fatalf("receiver.account_id = %v, want receiver-account-id", got)
	}
}

func TestPaymentDraftReceiverOmitsEmptyAccountId(t *testing.T) {
	body, err := json.Marshal(PaymentDraftReq{
		Title: "Payment draft",
		Payments: []PaymentDraftPayment{
			{
				Currency:  "GBP",
				Amount:    1234,
				AccountId: "source-account-id",
				Receiver: PaymentDraftPaymentReceiver{
					CounterpartyId: "counterparty-id",
				},
				Reference: "reference",
			},
		},
	})
	if err != nil {
		t.Fatalf("marshal payment draft request: %v", err)
	}

	var draft map[string]interface{}
	if err := json.Unmarshal(body, &draft); err != nil {
		t.Fatalf("unmarshal payment draft request: %v", err)
	}

	payments, ok := draft["payments"].([]interface{})
	if !ok || len(payments) != 1 {
		t.Fatalf("payments has unexpected value %T %v", draft["payments"], draft["payments"])
	}

	payment, ok := payments[0].(map[string]interface{})
	if !ok {
		t.Fatalf("payment has unexpected type %T", payments[0])
	}

	receiver, ok := payment["receiver"].(map[string]interface{})
	if !ok {
		t.Fatalf("receiver has unexpected type %T", payment["receiver"])
	}

	if _, ok := receiver["account_id"]; ok {
		t.Fatalf("draft receiver.account_id should be omitted when empty: %s", body)
	}
}
