package zkp

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	bn254_eddsa "github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/consensys/gnark-crypto/ecc/twistededwards"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/test"
	"github.com/Aliens0487/secFlow/pkg/circuits/statechecker"
	"github.com/Aliens0487/secFlow/pkg/crypto/keys"
	"github.com/Aliens0487/secFlow/pkg/model"
)

func (secFlow *SecFlowProgram) ComputeWitness(inputPath, keysPath, witnessFullPath, publicWitnessPath string) error {
	inputs, err := loadInputs(inputPath)
	if err != nil {
		return fmt.Errorf("error loading inputs: %w", err)
	}

	key, err := keys.LoadKeyPair(keysPath)
	if err != nil {
		return fmt.Errorf("error loading keys: %w", err)
	}
	keyPriv := key.(*bn254_eddsa.PrivateKey)

	var w statechecker.Circuit

	err = setupInputs(&w, inputs, keyPriv, secFlow)
	if err != nil {
		return fmt.Errorf("error setting up inputs: %w", err)
	}

	err = test.IsSolved(secFlow.Circuit, w, ecc.BN254.ScalarField())
	if err != nil {
		return fmt.Errorf("error checking if solved: %w", err)
	}

	log.Println("Creating witness")
	witnessFull, err := frontend.NewWitness(&w, ecc.BN254.ScalarField())
	if err != nil {
		return fmt.Errorf("error creating witness: %w", err)
	}

	log.Println("Writing witness to file")
	file, err := os.Create(witnessFullPath)
	if err != nil {
		return fmt.Errorf("error creating witness file: %w", err)
	}

	_, err = witnessFull.WriteTo(file)
	if err != nil {
		return fmt.Errorf("error writing witness to file: %w", err)
	}

	log.Println("Public witness")
	witnessPublic, err := frontend.NewWitness(&w, ecc.BN254.ScalarField(), frontend.PublicOnly())
	if err != nil {
		log.Fatalln("Failed to create public witness: ", err)
	}

	log.Println("Writing public witness to file")

	file, err = os.Create(publicWitnessPath)
	if err != nil {
		return fmt.Errorf("error creating public witness file: %w", err)
	}

	_, err = witnessPublic.WriteTo(file)
	if err != nil {
		return fmt.Errorf("error writing public witness to file: %w", err)
	}

	vector := witnessPublic.Vector().(fr.Vector)

	// Print public inputs
	// Public inputs: HashCurr, HashNew, PublicKey, Signature, Encrypted
	fmt.Println("Public inputs:", vector)
	return nil
}

func setupInputs(w *statechecker.Circuit, inputs *Inputs, keyPriv *bn254_eddsa.PrivateKey, secFlow *SecFlowProgram) error {
	w.State_curr = inputs.State_curr.toVariableState()
	w.State_new = inputs.State_new.toVariableState()

	w.HashCurr, _ = big.NewInt(0).SetString(inputs.HashCurr, 10)
	w.HashNew, _ = big.NewInt(0).SetString(inputs.HashNew, 10)
	w.Deposit = inputs.Deposit
	w.Withdrawal = inputs.Withdraw

	scalar := keys.GetPrivateKeyScaler(keyPriv)
	scalarHigh := new(big.Int).Rsh(scalar, 128)
	pow_2_128 := new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil)
	fmt.Println("pow_2_128:", pow_2_128.String())
	scalarLow := new(big.Int).Mod(scalar, pow_2_128)

	w.Keys.PrivateKey[0] = scalarHigh
	w.Keys.PrivateKey[1] = scalarLow
	fmt.Println("Private key:", scalar.String())
	pubKey := keyPriv.PublicKey
	w.Keys.PublicKey.A.X = pubKey.A.X
	w.Keys.PublicKey.A.Y = pubKey.A.Y

	xBytes := pubKey.A.X.Bytes()
	yBytes := pubKey.A.Y.Bytes()
	aBig := make([]*big.Int, 2)
	aBig[0] = new(big.Int).SetBytes(xBytes[:])
	aBig[1] = new(big.Int).SetBytes(yBytes[:])

	index := findParticipantIndex(secFlow.Model.Participants, [2]*big.Int(aBig))
	w.ParticipantIndex = index

	w.Encrypted = make([]frontend.Variable, len(inputs.Encrypted))
	for i, e := range inputs.Encrypted {
		w.Encrypted[i] = e
	}

	w.Key = make([]frontend.Variable, len(inputs.Key))
	for i, k := range inputs.Key {
		w.Key[i] = k
	}

	sigBytes, err := hex.DecodeString(inputs.Signature)
	if err != nil {
		return fmt.Errorf("error decoding signature: %w", err)
	}
	w.Signature.Assign(twistededwards.ID(ecc.BN254), sigBytes)

	w.Model = secFlow.Model
	w.VariableMapping = make(map[string]int)
	for i, n := range w.Model.Variables {
		w.VariableMapping[n] = i
	}
	return nil
}

func findParticipantIndex(participants []model.Participant, key [2]*big.Int) int {
	for i, p := range participants {
		if p.PublicKey[0].Cmp(key[0]) == 0 && p.PublicKey[1].Cmp(key[1]) == 0 {
			return i
		}
	}
	return -1
}
